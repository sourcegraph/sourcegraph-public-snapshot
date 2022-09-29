package dbstore

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

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
