package uploads

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Service) NewReferenceCountUpdater(interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.backfillReferenceCountBatch(ctx, batchSize)
	}))
}

func (s *Service) backfillReferenceCountBatch(ctx context.Context, batchSize int) (err error) {
	ctx, _, endObservation := s.operations.backfillReferenceCountBatch.With(ctx, &err, observation.Args{
		LogFields: []log.Field{log.Int("batchSize", batchSize)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.BackfillReferenceCountBatch(ctx, batchSize)
}
