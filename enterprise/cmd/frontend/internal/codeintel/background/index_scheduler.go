package background

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/autoindex/inference"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

type IndexScheduler struct {
	dbStore                     DBStore
	gitserverClient             GitserverClient
	batchSize                   int
	minimumTimeSinceLastEnqueue time.Duration
	minimumSearchCount          int
	minimumSearchRatio          float64
	minimumPreciseCount         int
	metrics                     Metrics
}

var _ goroutine.Handler = &IndexScheduler{}

func NewIndexScheduler(
	dbStore DBStore,
	gitserverClient GitserverClient,
	batchSize int,
	minimumTimeSinceLastEnqueue time.Duration,
	minimumSearchCount int,
	minimumSearchRatio float64,
	minimumPreciseCount int,
	interval time.Duration,
	metrics Metrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &IndexScheduler{
		dbStore:                     dbStore,
		gitserverClient:             gitserverClient,
		batchSize:                   batchSize,
		minimumTimeSinceLastEnqueue: minimumTimeSinceLastEnqueue,
		minimumSearchCount:          minimumSearchCount,
		minimumSearchRatio:          minimumSearchRatio,
		minimumPreciseCount:         minimumPreciseCount,
		metrics:                     metrics,
	})
}

func (s *IndexScheduler) Handle(ctx context.Context) error {
	configuredRepositoryIDs, err := s.dbStore.GetRepositoriesWithIndexConfiguration(ctx)
	if err != nil {
		return errors.Wrap(err, "store.GetRepositoriesWithIndexConfiguration")
	}

	indexableRepositories, err := s.dbStore.IndexableRepositories(ctx, store.IndexableRepositoryQueryOptions{
		Limit:                       s.batchSize,
		MinimumTimeSinceLastEnqueue: s.minimumTimeSinceLastEnqueue,
		MinimumSearchCount:          s.minimumSearchCount,
		MinimumPreciseCount:         s.minimumPreciseCount,
		MinimumSearchRatio:          s.minimumSearchRatio,
	})
	if err != nil {
		return errors.Wrap(err, "store.IndexableRepositories")
	}

	var indexableRepositoryIDs []int
	for _, indexableRepository := range indexableRepositories {
		indexableRepositoryIDs = append(indexableRepositoryIDs, indexableRepository.RepositoryID)
	}

	for _, repositoryID := range deduplicateRepositoryIDs(configuredRepositoryIDs, indexableRepositoryIDs) {
		if err := s.queueIndex(ctx, repositoryID); err != nil {
			if isRepoNotExist(err) {
				continue
			}

			return err
		}
	}

	return nil
}

func (s *IndexScheduler) HandleError(err error) {
	s.metrics.Errors.Inc()
	log15.Error("Failed to update indexable repositories", "err", err)
}

func (s *IndexScheduler) queueIndex(ctx context.Context, repositoryID int) (err error) {
	commit, err := s.gitserverClient.Head(ctx, s.dbStore, repositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	isQueued, err := s.dbStore.IsQueued(ctx, repositoryID, commit)
	if err != nil {
		return errors.Wrap(err, "store.IsQueued")
	}
	if isQueued {
		return nil
	}

	indexes, err := s.getIndexJobs(ctx, repositoryID, commit)
	if err != nil {
		return err
	}
	if len(indexes) == 0 {
		return nil
	}

	tx, err := s.dbStore.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "store.Transact")
	}
	defer func() {
		err = tx.Done(err)
	}()

	for _, index := range indexes {
		id, err := tx.InsertIndex(ctx, index)
		if err != nil {
			return errors.Wrap(err, "store.QueueIndex")
		}

		log15.Info(
			"Enqueued index",
			"id", id,
			"repository_id", repositoryID,
			"commit", commit,
		)
	}

	now := time.Now().UTC()
	update := store.UpdateableIndexableRepository{
		RepositoryID:        repositoryID,
		LastIndexEnqueuedAt: &now,
	}

	// TODO(efritz) - this may create records once a repository has an explicit
	// index configuration. This shouldn't affect any indexing behavior at all.
	if err := tx.UpdateIndexableRepository(ctx, update, now); err != nil {
		return errors.Wrap(err, "store.UpdateIndexableRepository")
	}

	return nil
}

func (s *IndexScheduler) getIndexJobs(ctx context.Context, repositoryID int, commit string) ([]store.Index, error) {
	fns := []func(ctx context.Context, repositoryID int, commit string) ([]store.Index, bool, error){
		s.getIndexJobsFromConfigurationInDatabase,
		s.getIndexJobsFromConfigurationInRepository,
		s.inferIndexJobsFromRepositoryStructure,
	}

	for _, fn := range fns {
		if indexJobs, ok, err := fn(ctx, repositoryID, commit); err != nil {
			return nil, err
		} else if ok {
			return indexJobs, nil
		}
	}

	return nil, nil
}

func (s *IndexScheduler) getIndexJobsFromConfigurationInDatabase(ctx context.Context, repositoryID int, commit string) ([]store.Index, bool, error) {
	indexConfigurationRecord, ok, err := s.dbStore.GetIndexConfigurationByRepositoryID(ctx, repositoryID)
	if err != nil {
		return nil, false, errors.Wrap(err, "store.GetIndexConfigurationByRepositoryID")
	}
	if !ok {
		return nil, false, nil
	}

	indexConfiguration, err := config.UnmarshalJSON(indexConfigurationRecord.Data)
	if err != nil {
		// We failed here, but do not try to fall back on another method as having
		// an explicit config in the database should always take precedence, even
		// if it's broken.
		log15.Warn("Failed to unmarshal index configuration", "repository_id", repositoryID)
		return nil, true, nil
	}

	return convertIndexConfiguration(repositoryID, commit, indexConfiguration), true, nil
}

func (s *IndexScheduler) getIndexJobsFromConfigurationInRepository(ctx context.Context, repositoryID int, commit string) ([]store.Index, bool, error) {
	isConfigured, err := s.gitserverClient.FileExists(ctx, s.dbStore, repositoryID, commit, "sourcegraph.yaml")
	if err != nil {
		return nil, false, errors.Wrap(err, "gitserver.FileExists")
	}
	if !isConfigured {
		return nil, false, nil
	}

	content, err := s.gitserverClient.RawContents(ctx, s.dbStore, repositoryID, commit, "sourcegraph.yaml")
	if err != nil {
		return nil, false, errors.Wrap(err, "gitserver.RawContents")
	}

	indexConfiguration, err := config.UnmarshalYAML(content)
	if err != nil {
		// We failed here, but do not try to fall back on another method as having
		// an explicit config in the repository should always take precedence over
		// an auto-inferred configuration, even if it's broken.
		log15.Warn("Failed to unmarshal index configuration", "repository_id", repositoryID)
		return nil, true, nil
	}

	return convertIndexConfiguration(repositoryID, commit, indexConfiguration), true, nil
}

func (s *IndexScheduler) inferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string) (indexes []store.Index, _ bool, _ error) {
	paths, err := s.gitserverClient.ListFiles(ctx, s.dbStore, repositoryID, commit, inference.Patterns)
	if err != nil {
		return nil, false, errors.Wrap(err, "gitserver.ListFiles")
	}

	for _, recognizer := range inference.Recognizers {
		indexes = append(indexes, convertInferredConfiguration(repositoryID, commit, recognizer.InferIndexJobs(paths))...)
	}

	return indexes, true, nil
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

func convertIndexConfiguration(repositoryID int, commit string, indexConfiguration config.IndexConfiguration) (indexes []store.Index) {
	for _, indexJob := range indexConfiguration.IndexJobs {
		var dockerSteps []store.DockerStep
		for _, dockerStep := range indexConfiguration.SharedSteps {
			dockerSteps = append(dockerSteps, store.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}
		for _, dockerStep := range indexJob.Steps {
			dockerSteps = append(dockerSteps, store.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}

		indexes = append(indexes, store.Index{
			Commit:       commit,
			RepositoryID: repositoryID,
			State:        "queued",
			DockerSteps:  dockerSteps,
			Root:         indexJob.Root,
			Indexer:      indexJob.Indexer,
			IndexerArgs:  indexJob.IndexerArgs,
			Outfile:      indexJob.Outfile,
		})
	}

	return indexes
}

func convertInferredConfiguration(repositoryID int, commit string, indexJobs []inference.IndexJob) (indexes []store.Index) {
	for _, indexJob := range indexJobs {
		var dockerSteps []store.DockerStep
		for _, dockerStep := range indexJob.DockerSteps {
			dockerSteps = append(dockerSteps, store.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}

		indexes = append(indexes, store.Index{
			RepositoryID: repositoryID,
			Commit:       commit,
			State:        "queued",
			DockerSteps:  dockerSteps,
			Root:         indexJob.Root,
			Indexer:      indexJob.Indexer,
			IndexerArgs:  indexJob.IndexerArgs,
			Outfile:      indexJob.Outfile,
		})
	}

	return indexes
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
