package store

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) ExpireFailedRecords(ctx context.Context, failedIndexMaxAge time.Duration, now time.Time) (err error) {
	ctx, _, endObservation := s.operations.setInferenceScript.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(expireFailedRecordsQuery, failedIndexMaxAge, now))
}

const expireFailedRecordsQuery = `
-- TODO
`
