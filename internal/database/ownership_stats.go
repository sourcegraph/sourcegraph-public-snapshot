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

// Note: owners FileWalkable passes in a map from string to int. The string is the owner in codeowners. If it's a handle, it has @ in front, otherwise email.
func (s *ownershipStats) UpdateCodeownersCounts(ctx context.Context, repoID api.RepoID, owners FileWalkable[map[string]int]) (int, error) {
	// TODO: Updates count
	codeownerIDs := map[string]int{} // cache fetched commit_author ids
	// Note: probably don't want use commit authors table here.
	return 0, owners.Walk(func(path string, ownedFileCountByOwner map[string]int) error {
		for identifier, count := range ownedFileCountByOwner {
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
		return nil
	})
}
