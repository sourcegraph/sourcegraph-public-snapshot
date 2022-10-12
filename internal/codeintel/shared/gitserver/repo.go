package gitserver

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type store struct {
	*basestore.Store
}

func newWithDB(db database.DB) *store {
	return &store{
		Store: basestore.NewWithHandle(db.Handle()),
	}
}

// ErrUnknownRepository occurs when a repository does not exist.
var ErrUnknownRepository = errors.New("unknown repository")

// RepoName returns the name for the repo with the given identifier.
func (s *store) RepoName(ctx context.Context, repositoryID int) (_ string, err error) {
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
SELECT name FROM repo WHERE id = %s
`

// RepoNames returns a map from repository id to names.
func (s *store) RepoNames(ctx context.Context, repositoryIDs ...int) (_ map[int]string, err error) {
	return scanRepoNames(s.Store.Query(ctx, sqlf.Sprintf(repoNamesQuery, pq.Array(repositoryIDs))))
}

const repoNamesQuery = `
SELECT id, name FROM repo WHERE id = ANY(%s)
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
