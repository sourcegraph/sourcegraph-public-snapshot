package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) RepoCount(ctx context.Context) (_ int, err error) {
	ctx, _, endObservation := s.operations.repoCount.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(repoCountQuery)))
	return count, err
}

const repoCountQuery = `
SELECT SUM(total)
FROM repo_statistics
`
