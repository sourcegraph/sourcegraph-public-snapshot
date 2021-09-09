package dbstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	Name                      string
	Type                      GitObjectType
	Pattern                   string
	RetentionEnabled          bool
	RetentionDuration         *time.Duration
	RetainIntermediateCommits bool
	IndexingEnabled           bool
	IndexCommitMaxAge         *time.Duration
	IndexIntermediateCommits  bool
}

// scanConfigurationPolicies scans a slice of configuration policies from the return value of `*Store.query`.
func scanConfigurationPolicies(rows *sql.Rows, queryErr error) (_ []ConfigurationPolicy, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var configurationPolicies []ConfigurationPolicy
	for rows.Next() {
		var configurationPolicy ConfigurationPolicy
		var retentionDurationHours, indexCommitMaxAgeHours *int

		if err := rows.Scan(
			&configurationPolicy.ID,
			&configurationPolicy.RepositoryID,
			&configurationPolicy.Name,
			&configurationPolicy.Type,
			&configurationPolicy.Pattern,
			&configurationPolicy.RetentionEnabled,
			&retentionDurationHours,
			&configurationPolicy.RetainIntermediateCommits,
			&configurationPolicy.IndexingEnabled,
			&indexCommitMaxAgeHours,
			&configurationPolicy.IndexIntermediateCommits,
		); err != nil {
			return nil, err
		}

		if retentionDurationHours != nil {
			duration := time.Duration(*retentionDurationHours) * time.Hour
			configurationPolicy.RetentionDuration = &duration
		}
		if indexCommitMaxAgeHours != nil {
			duration := time.Duration(*indexCommitMaxAgeHours) * time.Hour
			configurationPolicy.IndexCommitMaxAge = &duration
		}

		configurationPolicies = append(configurationPolicies, configurationPolicy)
	}

	return configurationPolicies, nil
}

// scanFirstConfigurationPolicy scans a slice of configuration policies from the return value of `*Store.query`
// and returns the first.
func scanFirstConfigurationPolicy(rows *sql.Rows, err error) (ConfigurationPolicy, bool, error) {
	scanConfigurationPolicies, err := scanConfigurationPolicies(rows, err)
	if err != nil || len(scanConfigurationPolicies) == 0 {
		return ConfigurationPolicy{}, false, err
	}
	return scanConfigurationPolicies[0], true, nil
}

type GetConfigurationPoliciesOptions struct {
	RepositoryID     int
	ForDataRetention bool
	ForIndexing      bool
}

// GetConfigurationPolicies retrieves the set of configuration policies matching the the given options.
func (s *Store) GetConfigurationPolicies(ctx context.Context, opts GetConfigurationPoliciesOptions) (_ []ConfigurationPolicy, err error) {
	ctx, traceLog, endObservation := s.operations.getConfigurationPolicies.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
	}})
	defer endObservation(1, observation.Args{})

	conds := make([]*sqlf.Query, 0, 3)
	if opts.RepositoryID == 0 {
		conds = append(conds, sqlf.Sprintf("repository_id IS NULL"))
	} else {
		conds = append(conds, sqlf.Sprintf("repository_id = %s", opts.RepositoryID))
	}
	if opts.ForDataRetention {
		conds = append(conds, sqlf.Sprintf("retention_enabled"))
	}
	if opts.ForIndexing {
		conds = append(conds, sqlf.Sprintf("indexing_enabled"))
	}

	configurationPolicies, err := scanConfigurationPolicies(s.Store.Query(ctx, sqlf.Sprintf(getConfigurationPoliciesQuery, sqlf.Join(conds, "AND"))))
	if err != nil {
		return nil, err
	}
	traceLog(log.Int("numConfigurationPolicies", len(configurationPolicies)))

	return configurationPolicies, nil
}

const getConfigurationPoliciesQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/configuration_policies.go:GetSomething
SELECT
	id,
	repository_id,
	name,
	type,
	pattern,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits
FROM lsif_configuration_policies
WHERE %s
ORDER BY name
`

// GetConfigurationPolicyByID retrieves the configuration policy with the given identifier.
func (s *Store) GetConfigurationPolicyByID(ctx context.Context, id int) (_ ConfigurationPolicy, _ bool, err error) {
	ctx, endObservation := s.operations.getConfigurationPolicyByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return scanFirstConfigurationPolicy(s.Store.Query(ctx, sqlf.Sprintf(getConfigurationPolicyByIDQuery, id)))
}

const getConfigurationPolicyByIDQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/configuration_policies.go:GetConfigurationPolicyByID
SELECT
	id,
	repository_id,
	name,
	type,
	pattern,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits
FROM lsif_configuration_policies
WHERE id = %s
`

// CreateConfigurationPolicy creates a configuration policy with the given fields (ignoring ID). The hydrated
// configuration policy record is returned.
func (s *Store) CreateConfigurationPolicy(ctx context.Context, configurationPolicy ConfigurationPolicy) (_ ConfigurationPolicy, err error) {
	ctx, endObservation := s.operations.createConfigurationPolicy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	var retentionDurationHours *int
	if configurationPolicy.RetentionDuration != nil {
		duration := int(*configurationPolicy.RetentionDuration / time.Hour)
		retentionDurationHours = &duration
	}

	var indexingCOmmitMaxAgeHours *int
	if configurationPolicy.IndexCommitMaxAge != nil {
		duration := int(*configurationPolicy.IndexCommitMaxAge / time.Hour)
		indexingCOmmitMaxAgeHours = &duration
	}

	hydratedConfigurationPolicy, _, err := scanFirstConfigurationPolicy(s.Query(ctx, sqlf.Sprintf(
		createConfigurationPolicyQuery,
		configurationPolicy.RepositoryID,
		configurationPolicy.Name,
		configurationPolicy.Type,
		configurationPolicy.Pattern,
		configurationPolicy.RetentionEnabled,
		retentionDurationHours,
		configurationPolicy.RetainIntermediateCommits,
		configurationPolicy.IndexingEnabled,
		indexingCOmmitMaxAgeHours,
		configurationPolicy.IndexIntermediateCommits,
	)))
	if err != nil {
		return ConfigurationPolicy{}, err
	}

	return hydratedConfigurationPolicy, nil
}

const createConfigurationPolicyQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/configuration_policies.go:CreateConfigurationPolicy
INSERT INTO lsif_configuration_policies (
	repository_id,
	name,
	type,
	pattern,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
RETURNING
	id,
	repository_id,
	name,
	type,
	pattern,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits
`

// UpdateConfigurationPolicy updates the fields of the configuration policy record with the given identifier.
func (s *Store) UpdateConfigurationPolicy(ctx context.Context, policy ConfigurationPolicy) (err error) {
	ctx, endObservation := s.operations.updateConfigurationPolicy.With(ctx, &err, observation.Args{LogFields: []log.Field{
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

	return s.Store.Exec(ctx, sqlf.Sprintf(updateConfigurationPolicyQuery,
		policy.Name,
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

const updateConfigurationPolicyQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/configuration_policies.go:UpdateConfigurationPolicy
UPDATE lsif_configuration_policies
SET
	name = %s,
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
	ctx, endObservation := s.operations.deleteConfigurationPolicyByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	return s.Store.Exec(ctx, sqlf.Sprintf(deleteConfigurationPolicyByIDQuery, id))
}

const deleteConfigurationPolicyByIDQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/configuration_policies.go:DeleteConfigurationPolicyByID
DELETE FROM lsif_configuration_policies WHERE id = %s
`
