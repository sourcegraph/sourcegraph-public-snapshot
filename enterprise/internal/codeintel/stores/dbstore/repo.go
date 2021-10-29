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

	return basestore.ScanInts(s.Store.Query(ctx, sqlf.Sprintf(repoIDsByGlobPatternQuery, makeWildcardPattern(pattern), authzConds)))
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

	return s.Store.Exec(ctx, sqlf.Sprintf(updateReposMatchingPatternsQuery, sqlf.Join(conds, "OR"), policyID, policyID))
}

const updateReposMatchingPatternsQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/repo.go:UpdateReposMatchingPatterns
WITH
	repos AS (
		SELECT id
		FROM repo
		WHERE
			(%s) AND
			deleted_at IS NULL AND
			blocked IS NULL
	),
	deleted AS (
		DELETE FROM lsif_configuration_policies_repository_pattern_lookup
		WHERE
			policy_id = %s AND
			-- Do not delete rows that we're inserting
			repo_id NOT IN (SELECT id FROM repos)
	)
INSERT INTO lsif_configuration_policies_repository_pattern_lookup(policy_id, repo_id)
SELECT %s, repos.id
FROM repos
`

func makeWildcardPattern(pattern string) string {
	return strings.ToLower(strings.ReplaceAll(pattern, "*", "%"))
}
