package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrUnknownRepository = errors.New("unknown repository")

func (s *store) GetRepoName(ctx context.Context, repositoryID int) (name string, err error) {
	ctx, _, endObservation := s.operations.getRepoName.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.String("repositoryName", name),
		}})
	}()

	name, exists, err := basestore.ScanFirstString(s.db.Query(ctx, sqlf.Sprintf(repoNameQuery, repositoryID)))
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
