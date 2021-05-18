package indexing

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

type IndexScheduler struct {
	dbStore                     DBStore
	indexEnqueuer               IndexEnqueuer
	operations                  *operations
	batchSize                   int
	minimumTimeSinceLastEnqueue time.Duration
	minimumSearchCount          int
	minimumSearchRatio          float64
	minimumPreciseCount         int
}

var _ goroutine.Handler = &IndexScheduler{}

func NewIndexScheduler(
	dbStore DBStore,
	indexEnqueuer IndexEnqueuer,
	batchSize int,
	minimumTimeSinceLastEnqueue time.Duration,
	minimumSearchCount int,
	minimumSearchRatio float64,
	minimumPreciseCount int,
	interval time.Duration,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	scheduler := &IndexScheduler{
		dbStore:                     dbStore,
		indexEnqueuer:               indexEnqueuer,
		batchSize:                   batchSize,
		minimumTimeSinceLastEnqueue: minimumTimeSinceLastEnqueue,
		minimumSearchCount:          minimumSearchCount,
		minimumSearchRatio:          minimumSearchRatio,
		minimumPreciseCount:         minimumPreciseCount,
		operations:                  newOperations(observationContext),
	}

	return goroutine.NewPeriodicGoroutineWithMetrics(
		context.Background(),
		interval,
		scheduler,
		scheduler.operations.HandleIndexScheduler,
	)
}

// For mocking in tests
var indexSchedulerEnabled = conf.CodeIntelAutoIndexingEnabled

func (s *IndexScheduler) Handle(ctx context.Context) error {
	if !indexSchedulerEnabled() {
		return nil
	}

	configuredRepositoryIDs, err := s.dbStore.GetRepositoriesWithIndexConfiguration(ctx)
	if err != nil {
		return errors.Wrap(err, "dbstore.GetRepositoriesWithIndexConfiguration")
	}

	indexableRepositories, err := s.dbStore.IndexableRepositories(ctx, store.IndexableRepositoryQueryOptions{
		Limit:                       s.batchSize,
		MinimumTimeSinceLastEnqueue: s.minimumTimeSinceLastEnqueue,
		MinimumSearchCount:          s.minimumSearchCount,
		MinimumPreciseCount:         s.minimumPreciseCount,
		MinimumSearchRatio:          s.minimumSearchRatio,
	})
	if err != nil {
		return errors.Wrap(err, "dbstore.IndexableRepositories")
	}

	var indexableRepositoryIDs []int
	for _, indexableRepository := range indexableRepositories {
		indexableRepositoryIDs = append(indexableRepositoryIDs, indexableRepository.RepositoryID)
	}

	var queueErr error
	for _, repositoryID := range deduplicateRepositoryIDs(configuredRepositoryIDs, indexableRepositoryIDs) {
		if err := s.indexEnqueuer.QueueIndexesForRepository(ctx, repositoryID); err != nil {
			if isRepoNotExist(err) {
				continue
			}

			if queueErr != nil {
				queueErr = err
			} else {
				queueErr = multierror.Append(queueErr, err)
			}
		}
	}
	if queueErr != nil {
		return queueErr
	}

	return nil
}

func (s *IndexScheduler) HandleError(err error) {
	log15.Error("Failed to update indexable repositories", "err", err)
}

func deduplicateRepositoryIDs(ids ...[]int) (repositoryIDs []int) {
	repositoryIDMap := map[int]struct{}{}
	for _, s := range ids {
		for _, v := range s {
			repositoryIDMap[v] = struct{}{}
		}
	}

	for repositoryID := range repositoryIDMap {
		repositoryIDs = append(repositoryIDs, repositoryID)
	}

	return repositoryIDs
}

func isRepoNotExist(err error) bool {
	for err != nil {
		if vcs.IsRepoNotExist(err) {
			return true
		}

		err = errors.Unwrap(err)
	}

	return false
}
