package store

import (
	"context"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetConfigurationPolicies retrieves the set of configuration policies matching the the given options.
// If a repository identifier is supplied (is non-zero), then only the configuration policies that apply
// to repository are returned. If repository is not supplied, then all policies may be returned.
func (s *store) GetConfigurationPolicies(ctx context.Context, opts shared.GetConfigurationPoliciesOptions) (_ []shared.ConfigurationPolicy, totalCount int, err error) {
	ctx, trace, endObservation := s.operations.getConfigurationPolicies.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
		log.String("term", opts.Term),
		log.Bool("forDataRetention", opts.ForDataRetention),
		log.Bool("forIndexing", opts.ForIndexing),
		log.Bool("forLockfileIndexing", opts.ForLockfileIndexing),
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
	if opts.ForLockfileIndexing {
		conds = append(conds, sqlf.Sprintf("p.lockfile_indexing_enabled"))
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
	trace.Log(log.Int("totalCount", totalCount))

	configurationPolicies, err := scanConfigurationPolicies(tx.Query(ctx, sqlf.Sprintf(
		getConfigurationPoliciesLimitedQuery,
		sqlf.Join(conds, "AND"),
		opts.Limit,
		opts.Offset,
	)))
	if err != nil {
		return nil, 0, err
	}
	trace.Log(log.Int("numConfigurationPolicies", len(configurationPolicies)))

	return configurationPolicies, totalCount, nil
}

const getConfigurationPoliciesCountQuery = `
-- source: internal/codeintel/policies/internal/store/store_configuration.go:GetConfigurationPolicies
SELECT COUNT(*)
FROM lsif_configuration_policies p
LEFT JOIN repo ON repo.id = p.repository_id
WHERE %s
`

const getConfigurationPoliciesUnlimitedQuery = `
-- source: internal/codeintel/policies/internal/store/store_configuration.go:GetConfigurationPolicies
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
	p.index_intermediate_commits,
	p.lockfile_indexing_enabled
FROM lsif_configuration_policies p
LEFT JOIN repo ON repo.id = p.repository_id
WHERE %s
ORDER BY p.name
`

const getConfigurationPoliciesLimitedQuery = getConfigurationPoliciesUnlimitedQuery + `
LIMIT %s OFFSET %s
`

// UpdateReposMatchingPatterns updates the values of the repository pattern lookup table for the
// given configuration policy identifier. Each repository matching one of the given patterns will
// be associated with the target configuration policy. If the patterns list is empty, the lookup
// table will remove any data associated with the target configuration policy. If the number of
// matches exceeds the given limit (if supplied), then only top ranked repositories by star count
// will be associated to the policy in the database and the remainder will be dropped.
func (s *store) UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) (err error) {
	ctx, _, endObservation := s.operations.updateReposMatchingPatterns.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("pattern", strings.Join(patterns, ",")),
		log.Int("policyID", policyID),
	}})
	defer endObservation(1, observation.Args{})

	n := len(patterns)
	if n == 0 {
		n = 1
	}

	conds := make([]*sqlf.Query, 0, n)
	for _, pattern := range patterns {
		conds = append(conds, sqlf.Sprintf("lower(name) LIKE %s", makeWildcardPattern(pattern)))
	}
	if len(patterns) == 0 {
		// We'll get a SQL syntax error if we try to join an empty disjunction list, so we
		// put this sentinel value here. Note that we choose FALSE over TRUE because we want
		// the absence of patterns to match NO repositories, not ALL repositories.
		conds = append(conds, sqlf.Sprintf("FALSE"))
	}

	limitExpression := sqlf.Sprintf("")
	if repositoryMatchLimit != nil {
		limitExpression = sqlf.Sprintf("LIMIT %s", *repositoryMatchLimit)
	}

	return s.db.Exec(ctx, sqlf.Sprintf(updateReposMatchingPatternsQuery, sqlf.Join(conds, "OR"), limitExpression, policyID, policyID, policyID))
}

const updateReposMatchingPatternsQuery = `
-- source: internal/codeintel/policies/internal/store/store_configuration.go:UpdateReposMatchingPatterns
WITH
matching_repositories AS (
	SELECT id AS repo_id
	FROM repo
	WHERE
		(%s) AND
		deleted_at IS NULL AND
		blocked IS NULL
	ORDER BY stars DESC NULLS LAST, id
	%s
),
inserted AS (
	-- Insert records that match the policy but don't yet exist
	INSERT INTO lsif_configuration_policies_repository_pattern_lookup(policy_id, repo_id)
	SELECT %s, r.repo_id
	FROM (
		SELECT r.repo_id
		FROM matching_repositories r
		WHERE r.repo_id NOT IN (
			SELECT repo_id
			FROM lsif_configuration_policies_repository_pattern_lookup
			WHERE policy_id = %s
		)
	) r
	ORDER BY r.repo_id
	RETURNING 1
),
locked_outdated_matching_repository_records AS (
	SELECT policy_id, repo_id
	FROM lsif_configuration_policies_repository_pattern_lookup
	WHERE
		policy_id = %s AND
		repo_id NOT IN (SELECT repo_id FROM matching_repositories)
	ORDER BY policy_id, repo_id FOR UPDATE
),
deleted AS (
	-- Delete records that no longer match the policy
	DELETE FROM lsif_configuration_policies_repository_pattern_lookup
	WHERE (policy_id, repo_id) IN (
		SELECT policy_id, repo_id
		FROM locked_outdated_matching_repository_records
	)
	RETURNING 1
)
SELECT
	(SELECT COUNT(*) FROM inserted) AS num_inserted,
	(SELECT COUNT(*) FROM deleted) AS num_deleted
`

func makeWildcardPattern(pattern string) string {
	return strings.ToLower(strings.ReplaceAll(pattern, "*", "%"))
}

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

// scanConfigurationPolicies scans a slice of configuration policies from the return value of `*Store.query`.
var scanConfigurationPolicies = basestore.NewSliceScanner(scanConfigurationPolicy)

func scanConfigurationPolicy(s dbutil.Scanner) (configurationPolicy shared.ConfigurationPolicy, err error) {
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
		&configurationPolicy.LockfileIndexingEnabled,
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
