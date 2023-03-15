package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"
	"go.opentelemetry.io/otel/attribute"

	policiesshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetConfigurationPolicies retrieves the set of configuration policies matching the the given options.
// If a repository identifier is supplied (is non-zero), then only the configuration policies that apply
// to repository are returned. If repository is not supplied, then all policies may be returned.
func (s *store) GetConfigurationPolicies(ctx context.Context, opts policiesshared.GetConfigurationPoliciesOptions) (_ []types.ConfigurationPolicy, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getConfigurationPolicies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
		log.String("term", opts.Term),
		log.Bool("forDataRetention", opts.ForDataRetention),
		log.Bool("forIndexing", opts.ForIndexing),
		log.Int("limit", opts.Limit),
		log.Int("offset", opts.Offset),
	}})
	defer endObservation(1, observation.Args{})

	conds := make([]*sqlf.Query, 0, 5)
	if opts.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf(`(
			(p.repository_id IS NULL AND p.repository_patterns IS NULL) OR
			p.repository_id = %s OR
			EXISTS (
				SELECT 1
				FROM lsif_configuration_policies_repository_pattern_lookup l
				WHERE l.policy_id = p.id AND l.repo_id = %s
			)
		)`, opts.RepositoryID, opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = append(conds, makeConfigurationPolicySearchCondition(opts.Term))
	}
	if opts.ForDataRetention {
		conds = append(conds, sqlf.Sprintf("p.retention_enabled"))
	}
	if opts.ForIndexing {
		conds = append(conds, sqlf.Sprintf("p.indexing_enabled"))
	}
	if opts.Protected != nil {
		if *opts.Protected {
			conds = append(conds, sqlf.Sprintf("p.protected"))
		} else {
			conds = append(conds, sqlf.Sprintf("NOT p.protected"))
		}
	}
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}

	tx, err := s.db.Transact(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = tx.Done(err) }()

	totalCount, _, err = basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(
		getConfigurationPoliciesCountQuery,
		sqlf.Join(conds, "AND"),
	)))
	if err != nil {
		return nil, 0, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("totalCount", totalCount))

	configurationPolicies, err := scanConfigurationPolicies(tx.Query(ctx, sqlf.Sprintf(
		getConfigurationPoliciesLimitedQuery,
		sqlf.Join(conds, "AND"),
		opts.Limit,
		opts.Offset,
	)))
	if err != nil {
		return nil, 0, err
	}
	trace.AddEvent("TODO Domain Owner", attribute.Int("numConfigurationPolicies", len(configurationPolicies)))

	return configurationPolicies, totalCount, nil
}

const getConfigurationPoliciesCountQuery = `
SELECT COUNT(*)
FROM lsif_configuration_policies p
LEFT JOIN repo ON repo.id = p.repository_id
WHERE %s
`

const getConfigurationPoliciesUnlimitedQuery = `
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
WHERE %s
ORDER BY p.name
`

const getConfigurationPoliciesLimitedQuery = getConfigurationPoliciesUnlimitedQuery + `
LIMIT %s OFFSET %s
`

// GetConfigurationPolicyByID retrieves the configuration policy with the given identifier.
func (s *store) GetConfigurationPolicyByID(ctx context.Context, id int) (_ types.ConfigurationPolicy, _ bool, err error) {
	ctx, _, endObservation := s.operations.getConfigurationPolicyByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return types.ConfigurationPolicy{}, false, err
	}

	return scanFirstConfigurationPolicy(s.db.Query(ctx, sqlf.Sprintf(getConfigurationPolicyByIDQuery, id, authzConds)))
}

const getConfigurationPolicyByIDQuery = `
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
func (s *store) CreateConfigurationPolicy(ctx context.Context, configurationPolicy types.ConfigurationPolicy) (_ types.ConfigurationPolicy, err error) {
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

	hydratedConfigurationPolicy, _, err := scanFirstConfigurationPolicy(s.db.Query(ctx, sqlf.Sprintf(
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
		return types.ConfigurationPolicy{}, err
	}

	return hydratedConfigurationPolicy, nil
}

const createConfigurationPolicyQuery = `
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
func (s *store) UpdateConfigurationPolicy(ctx context.Context, policy types.ConfigurationPolicy) (err error) {
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

	tx, err := s.db.Transact(ctx)
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
func (s *store) DeleteConfigurationPolicyByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.deleteConfigurationPolicyByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	protected, ok, err := basestore.ScanFirstBool(s.db.Query(ctx, sqlf.Sprintf(deleteConfigurationPolicyByIDQuery, id)))
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

// makeConfigurationPolicySearchCondition returns a disjunction of LIKE clauses against all searchable
// columns of an configuration policy.
func makeConfigurationPolicySearchCondition(term string) *sqlf.Query {
	searchableColumns := []string{
		"p.name",
	}

	var termConds []*sqlf.Query
	for _, column := range searchableColumns {
		termConds = append(termConds, sqlf.Sprintf(column+" ILIKE %s", "%"+term+"%"))
	}

	return sqlf.Sprintf("(%s)", sqlf.Join(termConds, " OR "))
}
