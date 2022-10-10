package ranking

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
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

	repos, err := s.store.GetRepos(ctx)
	if err != nil {
		return err
	}

	for _, repoName := range repos {
		if err := s.indexRepository(ctx, api.RepoName(repoName)); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) indexRepository(ctx context.Context, repoName api.RepoName) (err error) {
	_, _, endObservation := s.operations.indexRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// TODO
	ranks := map[string][]float64{
		string(repoName): {float64(len(repoName))},
		"bar":            {2.3},
	}

	return s.store.SetDocumentRanks(ctx, repoName, ranks)
}
