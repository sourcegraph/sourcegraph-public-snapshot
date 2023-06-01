package database

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OwnershipStatsStore interface {
	UpdateCodeownersCounts(ctx context.Context, repoID api.RepoID, owners FileWalkable[map[string]int]) (int, error)
	AggregateOwnership(ctx context.Context, opts OwnershipOpts, limitOffset *LimitOffset) ([]AggregateCodeowner, error)
	CodeownedFilesCount(ctx context.Context, opts OwnershipOpts) (int, error)
}

type ownershipStats struct {
	*basestore.Store
}

var ownerQueryFmtstr = `
	WITH existing (id) AS (
		SELECT a.id
		FROM commit_authors AS a
		WHERE a.email = %s
		AND a.name = %s
	), inserted (id) AS (
		INSERT INTO commit_authors (email, name)
		SELECT %s, %s
		WHERE NOT EXISTS (SELECT id FROM existing)
		RETURNING id
	)
	SELECT id FROM existing
	UNION ALL
	SELECT id FROM inserted
`

var upsertCount = `
	INSERT INTO codeowners_stats (file_path_id, codeowners_id, deep_file_count)
	SELECT p.id, %s, %s
	FROM repo_paths AS p
	WHERE p.repo_id = %s
	AND p.absolute_path = %s
	ON CONFLICT (file_path_id, codeowners_id)
	DO UPDATE
	SET deep_file_count = EXCLUDED.deep_file_count
`

var upsertTotalCount = `
INSERT INTO codeowners_aggregate_stats (file_path_id, deep_file_count)
SELECT p.id, %s
FROM repo_paths AS p
WHERE p.repo_id = %s
AND p.absolute_path = %s
ON CONFLICT (file_path_id)
DO UPDATE
SET deep_file_count = EXCLUDED.deep_file_count
`

// Note: owners FileWalkable passes in a map from string to int. The string is the owner in codeowners. If it's a handle, it has @ in front, otherwise email.
func (s *ownershipStats) UpdateCodeownersCounts(ctx context.Context, repoID api.RepoID, owners FileWalkable[map[string]int]) (int, error) {
	// TODO: Updates count
	codeownerIDs := map[string]int{} // cache fetched commit_author ids
	// Note: probably don't want use commit authors table here.
	return 0, owners.Walk(func(path string, ownedFileCountByOwner map[string]int) error {
		for identifier, count := range ownedFileCountByOwner {
			if identifier == "" {
				continue
			}
			var id int
			id = codeownerIDs[identifier]
			if id == 0 {
				var email, handle string
				if strings.HasPrefix(identifier, "@") {
					handle = strings.TrimPrefix(identifier, "@")
				} else {
					email = identifier
				}
				q := sqlf.Sprintf(ownerQueryFmtstr, email, handle, email, handle)
				r := s.Store.QueryRow(ctx, q)
				if err := r.Scan(&id); err != nil {
					return errors.Wrapf(err, q.Query(sqlf.PostgresBindVar))
				}
				codeownerIDs[identifier] = id
			}
			// At this point we assume paths exists in repo_paths, otherwise no update.
			q := sqlf.Sprintf(upsertCount, id, count, repoID, path)
			if err := s.Store.Exec(ctx, q); err != nil {
				return errors.Wrapf(err, q.Query(sqlf.PostgresBindVar))
			}
		}
		if totalFilesWithOwnership := ownedFileCountByOwner[""]; totalFilesWithOwnership > 0 {
			q := sqlf.Sprintf(upsertTotalCount, totalFilesWithOwnership, repoID, path)
			if err := s.Store.Exec(ctx, q); err != nil {
				return err
			}
		}
		return nil
	})
}

type OwnershipOpts struct {
	RepoID api.RepoID // Repo ID zero means all repos
	Path   string     // Path not set means repo root
}

type AggregateCodeowner struct {
	Handle    string
	Email     string
	FileCount int
}

var aggregateOwnershipFmtstr = `
	SELECT a.name, a.email, c.deep_file_count
	FROM codeowners_stats AS c
	INNER JOIN repo_paths AS p ON c.file_path_id = p.id
	INNER JOIN commit_authors AS a ON a.id = c.codeowners_id
	WHERE p.absolute_path = %s
`

func (s *ownershipStats) AggregateOwnership(ctx context.Context, opts OwnershipOpts, limitOffset *LimitOffset) ([]AggregateCodeowner, error) {
	q := []*sqlf.Query{sqlf.Sprintf(aggregateOwnershipFmtstr, opts.Path)}
	if repoID := opts.RepoID; repoID != 0 {
		q = append(q, sqlf.Sprintf("AND p.repo_id = %s", repoID))
	}
	q = append(q, sqlf.Sprintf("ORDER BY 3"))
	q = append(q, limitOffset.SQL())
	rs, err := s.Store.Query(ctx, sqlf.Join(q, "\n"))
	if err != nil {
		return nil, err
	}
	var owners []AggregateCodeowner
	for rs.Next() {
		var o AggregateCodeowner
		if err := rs.Scan(&o.Handle, &o.Email, &o.FileCount); err != nil {
			return nil, err
		}
		owners = append(owners, o)
	}
	return owners, nil
}

var codeownedFilesCountFmtstr = `
	SELECT SUM(s.deep_file_count)
	FROM codeowners_aggregate_stats AS s
	INNER JOIN repo_paths AS p ON s.file_path_id = p.id
	WHERE p.absolute_path = %s
`

func (s *ownershipStats) CodeownedFilesCount(ctx context.Context, opts OwnershipOpts) (int, error) {
	qs := []*sqlf.Query{sqlf.Sprintf(codeownedFilesCountFmtstr, opts.Path)}
	if repoID := opts.RepoID; repoID != 0 {
		qs = append(qs, sqlf.Sprintf("AND p.repo_id = %s", repoID))
	}
	q := sqlf.Join(qs, "\n")
	r := s.Store.QueryRow(ctx, q)
	var total int
	if err := r.Scan(&total); err != nil {
		return 0, errors.Wrapf(err, "Query: %s", q.Query(sqlf.PostgresBindVar))
	}
	return total, nil
}
