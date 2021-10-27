package dbstore

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// TODO - document
func (s *Store) FindRepos(ctx context.Context, pattern string) (_ []int, err error) {
	ctx, endObservation := s.operations.findRepos.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("pattern", pattern),
	}})
	defer endObservation(1, observation.Args{})

	return basestore.ScanInts(s.Store.Query(ctx, sqlf.Sprintf(findReposQuery, pattern)))
}

const updateReposMatchingPatterns = `
-- source: enterprise/internal/codeintel/stores/dbstore/repo.go:UpdateReposMatchingPatterns
WITH
repos AS (
    SELECT id
    FROM repo
    WHERE
		(%s)
      AND
	    deleted_at IS NULL AND
	    blocked IS NULL
),
deleted AS (
	DELETE FROM lsif_configuration_policies_repository_pattern_lookup
	WHERE policy_id = %s
)
INSERT INTO lsif_configuration_policies_repository_pattern_lookup(policy_id, repo_id)
SELECT %s, repos.id FROM repos
`

func (s *Store) UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int) (err error) {
	ctx, endObservation := s.operations.updateReposMatchingPatterns.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("pattern", strings.Join(patterns, ",")),
	}})
	defer endObservation(1, observation.Args{})

	conds := make([]*sqlf.Query, 0, len(patterns))
	if len(patterns) == 0 {
		// When patterns is zero, we set the WHERE clause to FALSE
		// to make sure `repos` is empty so we can just trigger the `deleted` CTE.
		conds = append(conds, sqlf.Sprintf("FALSE"))
	} else {
		for _, pattern := range patterns {
			conds = append(conds, sqlf.Sprintf("name ILIKE %s", pattern))
		}
	}


	return s.Store.Exec(ctx, sqlf.Sprintf(updateReposMatchingPatterns, sqlf.Join(conds, "OR"), policyID, policyID))
}

//
// TODO - authz filters
//

const findReposQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/repo.go:FindRepos
SELECT id
FROM repo
WHERE
	name ILIKE %s AND
	deleted_at IS NULL AND
	blocked IS NULL
`
