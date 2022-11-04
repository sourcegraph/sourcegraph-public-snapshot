package dbstore

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitObjectType string

const (
	GitObjectTypeCommit GitObjectType = "GIT_COMMIT"
	GitObjectTypeTag    GitObjectType = "GIT_TAG"
	GitObjectTypeTree   GitObjectType = "GIT_TREE"
)

type ConfigurationPolicy struct {
	ID                        int
	RepositoryID              *int
	RepositoryPatterns        *[]string
	Name                      string
	Type                      GitObjectType
	Pattern                   string
	Protected                 bool
	RetentionEnabled          bool
	RetentionDuration         *time.Duration
	RetainIntermediateCommits bool
	IndexingEnabled           bool
	IndexCommitMaxAge         *time.Duration
	IndexIntermediateCommits  bool
}

func scanConfigurationPolicy(s dbutil.Scanner) (configurationPolicy ConfigurationPolicy, err error) {
	var repositoryPatterns []string
	var retentionDurationHours, indexCommitMaxAgeHours *int

	if err := s.Scan(
		&configurationPolicy.ID,
		&configurationPolicy.RepositoryID,
		pq.Array(&repositoryPatterns),
		&configurationPolicy.Name,
		&configurationPolicy.Type,
		&configurationPolicy.Pattern,
		&configurationPolicy.Protected,
		&configurationPolicy.RetentionEnabled,
		&retentionDurationHours,
		&configurationPolicy.RetainIntermediateCommits,
		&configurationPolicy.IndexingEnabled,
		&indexCommitMaxAgeHours,
		&configurationPolicy.IndexIntermediateCommits,
	); err != nil {
		return configurationPolicy, err
	}

	if len(repositoryPatterns) != 0 {
		configurationPolicy.RepositoryPatterns = &repositoryPatterns
	}
	if retentionDurationHours != nil {
		duration := time.Duration(*retentionDurationHours) * time.Hour
		configurationPolicy.RetentionDuration = &duration
	}
	if indexCommitMaxAgeHours != nil {
		duration := time.Duration(*indexCommitMaxAgeHours) * time.Hour
		configurationPolicy.IndexCommitMaxAge = &duration
	}
	return configurationPolicy, nil
}

// scanConfigurationPolicies scans a slice of configuration policies from the return value of `*Store.query`.
var scanConfigurationPolicies = basestore.NewSliceScanner(scanConfigurationPolicy)

// scanFirstConfigurationPolicy scans a slice of configuration policies from the return value of `*Store.query`
// and returns the first.
var scanFirstConfigurationPolicy = basestore.NewFirstScanner(scanConfigurationPolicy)

type GetConfigurationPoliciesOptions struct {
	// RepositoryID indicates that only configuration policies that apply to the
	// specified repository (directly or via pattern) should be returned. This value
	// has no effect when equal to zero.
	RepositoryID int

	// Term is a string to search within the configuration title.
	Term string

	// ForIndexing indicates that only configuration policies with data retention enabled
	// should be returned.
	ForDataRetention bool

	// ForIndexing indicates that only configuration policies with indexing enabled should
	// be returned.
	ForIndexing bool

	// Limit indicates the number of results to take from the result set.
	Limit int

	// Offset indicates the number of results to skip in the result set.
	Offset int
}

// GetConfigurationPolicyByID retrieves the configuration policy with the given identifier.
func (s *Store) GetConfigurationPolicyByID(ctx context.Context, id int) (_ ConfigurationPolicy, _ bool, err error) {
	ctx, _, endObservation := s.operations.getConfigurationPolicyByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.Store))
	if err != nil {
		return ConfigurationPolicy{}, false, err
	}

	return scanFirstConfigurationPolicy(s.Store.Query(ctx, sqlf.Sprintf(getConfigurationPolicyByIDQuery, id, authzConds)))
}

const getConfigurationPolicyByIDQuery = `
-- source: internal/codeintel/stores/dbstore/configuration_policies.go:GetConfigurationPolicyByID
SELECT
	p.id,
	p.repository_id,
	p.repository_patterns,
	p.name,
	p.type,
	p.pattern,
	p.protected,
	p.retention_enabled,
	p.retention_duration_hours,
	p.retain_intermediate_commits,
	p.indexing_enabled,
	p.index_commit_max_age_hours,
	p.index_intermediate_commits
FROM lsif_configuration_policies p
LEFT JOIN repo ON repo.id = p.repository_id
-- Global policies are visible to anyone
-- Repository-specific policies must check repository permissions
WHERE p.id = %s AND (p.repository_id IS NULL OR (%s))
`

// CreateConfigurationPolicy creates a configuration policy with the given fields (ignoring ID). The hydrated
// configuration policy record is returned.
func (s *Store) CreateConfigurationPolicy(ctx context.Context, configurationPolicy ConfigurationPolicy) (_ ConfigurationPolicy, err error) {
	ctx, _, endObservation := s.operations.createConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var retentionDurationHours *int
	if configurationPolicy.RetentionDuration != nil {
		duration := int(*configurationPolicy.RetentionDuration / time.Hour)
		retentionDurationHours = &duration
	}

	var indexingCommitMaxAgeHours *int
	if configurationPolicy.IndexCommitMaxAge != nil {
		duration := int(*configurationPolicy.IndexCommitMaxAge / time.Hour)
		indexingCommitMaxAgeHours = &duration
	}

	var repositoryPatterns any
	if configurationPolicy.RepositoryPatterns != nil {
		repositoryPatterns = pq.Array(*configurationPolicy.RepositoryPatterns)
	}

	hydratedConfigurationPolicy, _, err := scanFirstConfigurationPolicy(s.Query(ctx, sqlf.Sprintf(
		createConfigurationPolicyQuery,
		configurationPolicy.RepositoryID,
		repositoryPatterns,
		configurationPolicy.Name,
		configurationPolicy.Type,
		configurationPolicy.Pattern,
		configurationPolicy.RetentionEnabled,
		retentionDurationHours,
		configurationPolicy.RetainIntermediateCommits,
		configurationPolicy.IndexingEnabled,
		indexingCommitMaxAgeHours,
		configurationPolicy.IndexIntermediateCommits,
	)))
	if err != nil {
		return ConfigurationPolicy{}, err
	}

	return hydratedConfigurationPolicy, nil
}

const createConfigurationPolicyQuery = `
-- source: internal/codeintel/stores/dbstore/configuration_policies.go:CreateConfigurationPolicy
INSERT INTO lsif_configuration_policies (
	repository_id,
	repository_patterns,
	name,
	type,
	pattern,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
	id,
	repository_id,
	repository_patterns,
	name,
	type,
	pattern,
	false as protected,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits
`

var (
	errUnknownConfigurationPolicy       = errors.New("unknown configuration policy")
	errIllegalConfigurationPolicyUpdate = errors.New("protected configuration policies must keep the same names, types, patterns, and retention values (except duration)")
	errIllegalConfigurationPolicyDelete = errors.New("protected configuration policies cannot be deleted")
)

// UpdateConfigurationPolicy updates the fields of the configuration policy record with the given identifier.
func (s *Store) UpdateConfigurationPolicy(ctx context.Context, policy ConfigurationPolicy) (err error) {
	ctx, _, endObservation := s.operations.updateConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", policy.ID),
	}})
	defer endObservation(1, observation.Args{})

	var retentionDuration *int
	if policy.RetentionDuration != nil {
		duration := int(*policy.RetentionDuration / time.Hour)
		retentionDuration = &duration
	}

	var indexCommitMaxAge *int
	if policy.IndexCommitMaxAge != nil {
		duration := int(*policy.IndexCommitMaxAge / time.Hour)
		indexCommitMaxAge = &duration
	}

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// First, pull current policy to see if it's protected, and if so whether or not the
	// fields that must remain stable (names, types, patterns, and retention enabled) have
	// the same current and target values.

	currentPolicy, ok, err := scanFirstConfigurationPolicy(tx.Query(ctx, sqlf.Sprintf(updateConfigurationPolicySelectQuery, policy.ID)))
	if err != nil {
		return err
	}
	if !ok {
		return errUnknownConfigurationPolicy
	}
	if currentPolicy.Protected {
		if policy.Name != currentPolicy.Name || policy.Type != currentPolicy.Type || policy.Pattern != currentPolicy.Pattern || policy.RetentionEnabled != currentPolicy.RetentionEnabled || policy.RetainIntermediateCommits != currentPolicy.RetainIntermediateCommits {
			return errIllegalConfigurationPolicyUpdate
		}
	}

	var repositoryPatterns any
	if policy.RepositoryPatterns != nil {
		repositoryPatterns = pq.Array(*policy.RepositoryPatterns)
	}

	return tx.Exec(ctx, sqlf.Sprintf(updateConfigurationPolicyQuery,
		policy.Name,
		repositoryPatterns,
		policy.Type,
		policy.Pattern,
		policy.RetentionEnabled,
		retentionDuration,
		policy.RetainIntermediateCommits,
		policy.IndexingEnabled,
		indexCommitMaxAge,
		policy.IndexIntermediateCommits,
		policy.ID,
	))
}

const updateConfigurationPolicySelectQuery = `
-- source: internal/codeintel/stores/dbstore/configuration_policies.go:UpdateConfigurationPolicy
SELECT
	id,
	repository_id,
	repository_patterns,
	name,
	type,
	pattern,
	protected,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits
FROM lsif_configuration_policies
WHERE id = %s
FOR UPDATE
`

const updateConfigurationPolicyQuery = `
-- source: internal/codeintel/stores/dbstore/configuration_policies.go:UpdateConfigurationPolicy
UPDATE lsif_configuration_policies SET
	name = %s,
	repository_patterns = %s,
	type = %s,
	pattern = %s,
	retention_enabled = %s,
	retention_duration_hours = %s,
	retain_intermediate_commits = %s,
	indexing_enabled = %s,
	index_commit_max_age_hours = %s,
	index_intermediate_commits = %s
WHERE id = %s
`

// DeleteConfigurationPolicyByID deletes the configuration policy with the given identifier.
func (s *Store) DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.deleteConfigurationPolicyByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	protected, ok, err := basestore.ScanFirstBool(s.Store.Query(ctx, sqlf.Sprintf(deleteConfigurationPolicyByIDQuery, id)))
	if err != nil {
		return err
	}
	if !ok {
		return errUnknownConfigurationPolicy
	}
	if protected {
		return errIllegalConfigurationPolicyDelete
	}

	return nil
}

const deleteConfigurationPolicyByIDQuery = `
-- source: internal/codeintel/stores/dbstore/configuration_policies.go:DeleteConfigurationPolicyByID
WITH
candidate AS (
	SELECT id, protected
	FROM lsif_configuration_policies
	WHERE id = %s
	ORDER BY id FOR UPDATE
),
deleted AS (
	DELETE FROM lsif_configuration_policies WHERE id IN (SELECT id FROM candidate WHERE NOT protected)
)
SELECT protected FROM candidate
`

// SelectPoliciesForRepositoryMembershipUpdate returns a slice of configuration policies that should be considered
// for repository membership updates. Configuration policies are returned in the order of least recently updated.
func (s *Store) SelectPoliciesForRepositoryMembershipUpdate(ctx context.Context, batchSize int) (configurationPolicies []ConfigurationPolicy, err error) {
	ctx, trace, endObservation := s.operations.selectPoliciesForRepositoryMembershipUpdate.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	configurationPolicies, err = scanConfigurationPolicies(s.Store.Query(ctx, sqlf.Sprintf(selectPoliciesForRepositoryMembershipUpdate, batchSize, timeutil.Now())))
	if err != nil {
		return nil, err
	}
	trace.Log(log.Int("numConfigurationPolicies", len(configurationPolicies)))

	return configurationPolicies, nil
}

const selectPoliciesForRepositoryMembershipUpdate = `
-- source: internal/codeintel/stores/dbstore/configuration_policies.go:SelectPoliciesForRepositoryMembershipUpdate
WITH
candidate_policies AS (
	SELECT p.id
	FROM lsif_configuration_policies p
	ORDER BY p.last_resolved_at NULLS FIRST
	LIMIT %d
),
locked_policies AS (
	SELECT c.id
	FROM candidate_policies c
	ORDER BY c.id FOR UPDATE
)
UPDATE lsif_configuration_policies
SET last_resolved_at = %s
WHERE id IN (SELECT id FROM locked_policies)
RETURNING
	id,
	repository_id,
	repository_patterns,
	name,
	type,
	pattern,
	protected,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits
`
