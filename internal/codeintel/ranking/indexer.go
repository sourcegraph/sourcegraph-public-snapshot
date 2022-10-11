package ranking

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Service) RepositoryIndexer(interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.indexRepositories(ctx)
	}))
}

var rankingEnabled, _ = strconv.ParseBool(os.Getenv("ENABLE_EXPERIMENTAL_RANKING"))

func (s *Service) indexRepositories(ctx context.Context) (err error) {
	if !rankingEnabled {
		return nil
	}

	_, _, endObservation := s.operations.indexRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return errors.New("codeintel.ranking.service.indexRepositories unimplemented")
}
