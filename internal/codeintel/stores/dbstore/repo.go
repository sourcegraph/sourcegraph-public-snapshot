package dbstore

import (
	"context"
	"database/sql"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrUnknownRepository occurs when a repository does not exist.
var ErrUnknownRepository = errors.New("unknown repository")

// RepoName returns the name for the repo with the given identifier.
func (s *Store) RepoName(ctx context.Context, repositoryID int) (_ string, err error) {
	ctx, _, endObservation := s.operations.repoName.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	name, exists, err := basestore.ScanFirstString(s.Store.Query(ctx, sqlf.Sprintf(repoNameQuery, repositoryID)))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrUnknownRepository
	}
	return name, nil
}

const repoNameQuery = `
-- source: internal/codeintel/stores/dbstore/repos.go:RepoName
SELECT name FROM repo WHERE id = %s
`

func scanRepoNames(rows *sql.Rows, queryErr error) (_ map[int]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	names := map[int]string{}

	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}

		names[id] = name
	}

	return names, nil
}

// RepoNames returns a map from repository id to names.
func (s *Store) RepoNames(ctx context.Context, repositoryIDs ...int) (_ map[int]string, err error) {
	ctx, _, endObservation := s.operations.repoName.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numRepositories", len(repositoryIDs)),
	}})
	defer endObservation(1, observation.Args{})

	return scanRepoNames(s.Store.Query(ctx, sqlf.Sprintf(repoNamesQuery, pq.Array(repositoryIDs))))
}

const repoNamesQuery = `
-- source: internal/codeintel/stores/dbstore/repos.go:RepoNames
SELECT id, name FROM repo WHERE id = ANY(%s)
`

// RepoIDsByGlobPatterns returns a page of repository identifiers and a total count of repositories matching
// one of the given patterns.
func (s *Store) RepoIDsByGlobPatterns(ctx context.Context, patterns []string, limit, offset int) (_ []int, _ int, err error) {
	ctx, _, endObservation := s.operations.repoIDsByGlobPatterns.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("patterns", strings.Join(patterns, ", ")),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	if len(patterns) == 0 {
		return nil, 0, nil
	}

	conds := make([]*sqlf.Query, 0, len(patterns))
	for _, pattern := range patterns {
		conds = append(conds, sqlf.Sprintf("lower(name) LIKE %s", makeWildcardPattern(pattern)))
	}

	tx, err := s.transact(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer func() { err = tx.Done(err) }()

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDB(tx.Handle().DB()))
	if err != nil {
		return nil, 0, err
	}

	totalCount, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf(repoIDsByGlobPatternsCountQuery, sqlf.Join(conds, "OR"), authzConds)))
	if err != nil {
		return nil, 0, err
	}

	ids, err := basestore.ScanInts(tx.Query(ctx, sqlf.Sprintf(repoIDsByGlobPatternsQuery, sqlf.Join(conds, "OR"), authzConds, limit, offset)))
	if err != nil {
		return nil, 0, err
	}

	return ids, totalCount, nil
}

const repoIDsByGlobPatternsCountQuery = `
-- source: internal/codeintel/stores/dbstore/repo.go:RepoIDsByGlobPatterns
SELECT COUNT(*)
FROM repo
WHERE
	(%s) AND
	deleted_at IS NULL AND
	blocked IS NULL AND
	(%s)
`

const repoIDsByGlobPatternsQuery = `
-- source: internal/codeintel/stores/dbstore/repo.go:RepoIDsByGlobPatterns
SELECT id
FROM repo
WHERE
	(%s) AND
	deleted_at IS NULL AND
	blocked IS NULL AND
	(%s)
ORDER BY stars DESC NULLS LAST, id
LIMIT %s OFFSET %s
`

// UpdateReposMatchingPatterns updates the values of the repository pattern lookup table for the
// given configuration policy identifier. Each repository matching one of the given patterns will
// be associated with the target configuration policy. If the patterns list is empty, the lookup
// table will remove any data associated with the target configuration policy. If the number of
// matches exceeds the given limit (if supplied), then only top ranked repositories by star count
// will be associated to the policy in the database and the remainder will be dropped.
func (s *Store) UpdateReposMatchingPatterns(ctx context.Context, patterns []string, policyID int, repositoryMatchLimit *int) (err error) {
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

	return s.Store.Exec(ctx, sqlf.Sprintf(updateReposMatchingPatternsQuery, sqlf.Join(conds, "OR"), limitExpression, policyID, policyID, policyID))
}

const updateReposMatchingPatternsQuery = `
-- source: internal/codeintel/stores/dbstore/repo.go:UpdateReposMatchingPatterns
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
