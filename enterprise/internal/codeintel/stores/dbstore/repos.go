package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// RepoName returns the name for the repo with the given identifier.
func (s *Store) RepoName(ctx context.Context, repositoryID int) (_ string, err error) {
	ctx, endObservation := s.operations.repoName.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	name, exists, err := basestore.ScanFirstString(s.Store.Query(
		ctx,
		sqlf.Sprintf(`SELECT name FROM repo WHERE id = %s`, repositoryID),
	))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrUnknownRepository
	}
	return name, nil
}
