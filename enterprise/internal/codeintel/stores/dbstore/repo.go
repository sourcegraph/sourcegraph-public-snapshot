package dbstore

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// PolicyRepositoryMatchLimit is the maximum number of repositories we'll allow a single configuration
// policy to match. If there are more than this number of repositories, ones with a GitHub higher star
// count will be applied over those with a lower star count (to maintain relevance in Cloud).
//
// See #26852 for improvement ideas.
const PolicyRepositoryMatchLimit = 10000

// RepoIDsByGlobPattern returns a slice of IDs from the repo table that matches the pattern string.
func (s *Store) RepoIDsByGlobPattern(ctx context.Context, pattern string) (_ []int, err error) {
	ctx, endObservation := s.operations.repoIDsByGlobPattern.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("pattern", pattern),
	}})
	defer endObservation(1, observation.Args{})

	authzConds, err := database.AuthzQueryConds(ctx, s.Store.Handle().DB())
	if err != nil {
		return nil, err
	}

	return basestore.ScanInts(s.Store.Query(ctx, sqlf.Sprintf(repoIDsByGlobPatternQuery, makeWildcardPattern(pattern), authzConds, PolicyRepositoryMatchLimit)))
}

const repoIDsByGlobPatternQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/repo.go:RepoIDsByGlobPattern
SELECT id
FROM repo
WHERE
	lower(name) LIKE %s AND
	deleted_at IS NULL AND
	blocked IS NULL AND
	(%s)
ORDER BY stars DESC, id
LIMIT %s
`

// UpdateReposMatchingPatterns updates the values of the repository pattern lookup table for the
// given configuration policy identifier. Each repository matching one of the given patterns will
// be associated with the target configuration policy. If the patterns list is empty, the lookup
// table will remove any data associated with the target configuration policy.
func (s *Store) UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int) (err error) {
	ctx, endObservation := s.operations.updateReposMatchingPatterns.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("pattern", strings.Join(patterns, ",")),
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

	return s.Store.Exec(ctx, sqlf.Sprintf(updateReposMatchingPatternsQuery, sqlf.Join(conds, "OR"), PolicyRepositoryMatchLimit, policyID, policyID, policyID))
}

const updateReposMatchingPatternsQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/repo.go:UpdateReposMatchingPatterns
WITH
matching_repositories AS (
	SELECT id AS repo_id
	FROM repo
	WHERE
		(%s) AND
		deleted_at IS NULL AND
		blocked IS NULL
	ORDER BY stars DESC, id
	LIMIT %s
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
