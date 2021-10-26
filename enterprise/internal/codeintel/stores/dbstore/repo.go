package dbstore

import (
	"context"

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
