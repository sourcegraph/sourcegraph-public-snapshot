package scheduler

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	indexpkg "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/index"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

type Scheduler struct {
	store                       store.Store
	gitserverClient             gitserver.Client
	batchSize                   int
	minimumTimeSinceLastEnqueue time.Duration
	minimumSearchCount          int
	minimumSearchRatio          float64
	minimumPreciseCount         int
	metrics                     SchedulerMetrics
}

var _ goroutine.Handler = &Scheduler{}

func NewScheduler(
	store store.Store,
	gitserverClient gitserver.Client,
	interval time.Duration,
	batchSize int,
	minimumTimeSinceLastEnqueue time.Duration,
	minimumSearchCount int,
	minimumSearchRatio float64,
	minimumPreciseCount int,
	metrics SchedulerMetrics,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &Scheduler{
		store:                       store,
		gitserverClient:             gitserverClient,
		batchSize:                   batchSize,
		minimumTimeSinceLastEnqueue: minimumTimeSinceLastEnqueue,
		minimumSearchCount:          minimumSearchCount,
		minimumSearchRatio:          minimumSearchRatio,
		minimumPreciseCount:         minimumPreciseCount,
		metrics:                     metrics,
	})
}

func (s *Scheduler) Handle(ctx context.Context) error {
	indexableRepositories, err := s.store.IndexableRepositories(ctx, store.IndexableRepositoryQueryOptions{
		Limit:                       s.batchSize,
		MinimumTimeSinceLastEnqueue: s.minimumTimeSinceLastEnqueue,
		MinimumSearchCount:          s.minimumSearchCount,
		MinimumPreciseCount:         s.minimumPreciseCount,
		MinimumSearchRatio:          s.minimumSearchRatio,
	})
	if err != nil {
		return errors.Wrap(err, "store.IndexableRepositories")
	}

	for _, indexableRepository := range indexableRepositories {
		if err := s.queueIndex(ctx, indexableRepository); err != nil {
			if isRepoNotExist(err) {
				continue
			}

			return err
		}
	}

	return nil
}

func (s *Scheduler) HandleError(err error) {
	s.metrics.Errors.Inc()
	log15.Error("Failed to update indexable repositories", "err", err)
}

func (s *Scheduler) queueIndex(ctx context.Context, indexableRepository store.IndexableRepository) (err error) {
	commit, err := s.gitserverClient.Head(ctx, s.store, indexableRepository.RepositoryID)
	if err != nil {
		return errors.Wrap(err, "gitserver.Head")
	}

	isQueued, err := s.store.IsQueued(ctx, indexableRepository.RepositoryID, commit) // TODO - expand this check?
	if err != nil {
		return errors.Wrap(err, "store.IsQueued")
	}
	if isQueued {
		return nil
	}

	indexes, err := s.getIndexes(ctx, indexableRepository, commit)
	if err != nil {
		return err
	}

	tx, err := s.store.Transact(ctx)
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
			"repository_id", indexableRepository.RepositoryID,
			"commit", commit,
		)

	}

	now := time.Now().UTC()
	update := store.UpdateableIndexableRepository{
		RepositoryID:        indexableRepository.RepositoryID,
		LastIndexEnqueuedAt: &now,
	}

	if err := tx.UpdateIndexableRepository(ctx, update, now); err != nil {
		return errors.Wrap(err, "store.UpdateIndexableRepository")
	}

	return nil
}

func (s *Scheduler) getIndexes(ctx context.Context, indexableRepository store.IndexableRepository, commit string) (indexes []store.Index, _ error) {
	isConfigured, err := s.gitserverClient.FileExists(ctx, s.store, indexableRepository.RepositoryID, commit, "sourcegraph.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.FileExists")
	}

	if isConfigured {
		return s.indexesFromConfiguration(ctx, indexableRepository, commit)
	}

	return s.inferIndexes(ctx, indexableRepository, commit)
}

func (s *Scheduler) indexesFromConfiguration(ctx context.Context, indexableRepository store.IndexableRepository, commit string) (indexes []store.Index, _ error) {
	content, err := s.gitserverClient.Text(ctx, s.store, indexableRepository.RepositoryID, commit, "sourcegraph.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.Text")
	}

	configuration, err := indexpkg.UnmarshalYAML(content)
	if err != nil {
		return nil, errors.Wrap(err, "index.UnmarshalYAML")
	}

	for _, indexJob := range configuration.IndexJobs {
		dockerSteps := []store.DockerStep{}
		for _, step := range configuration.SharedSteps {
			dockerSteps = append(dockerSteps, store.DockerStep{
				Root:     step.Root,
				Image:    step.Image,
				Commands: step.Commands,
			})
		}
		for _, step := range indexJob.Steps {
			dockerSteps = append(dockerSteps, store.DockerStep{
				Root:     step.Root,
				Image:    step.Image,
				Commands: step.Commands,
			})
		}

		indexes = append(indexes, store.Index{
			State:        "queued",
			Commit:       commit,
			RepositoryID: indexableRepository.RepositoryID,
			DockerSteps:  dockerSteps,
			Root:         indexJob.Root,
			Indexer:      indexJob.Indexer,
			IndexerArgs:  indexJob.IndexerArgs,
			Outfile:      indexJob.Outfile,
		})
	}

	return indexes, nil
}

func (s *Scheduler) inferIndexes(ctx context.Context, indexableRepository store.IndexableRepository, commit string) ([]store.Index, error) {
	isGo, err := s.gitserverClient.FileExists(ctx, s.store, indexableRepository.RepositoryID, commit, "go.mod")
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.FileExists")
	}

	if isGo {
		return []store.Index{
			{
				State:        "queued",
				Commit:       commit,
				RepositoryID: indexableRepository.RepositoryID,
				DockerSteps:  []store.DockerStep{},
				Root:         "", // TODO
				Indexer:      "sourcegraph/lsif-go:latest",
				IndexerArgs:  []string{"lsif-go", "--no-animation"},
				Outfile:      "",
			},
		}, nil
	}

	return []store.Index{
		{
			State:        "queued",
			Commit:       commit,
			RepositoryID: indexableRepository.RepositoryID,
			DockerSteps: []store.DockerStep{
				{
					Root:     "", // TODO
					Image:    "circleci/node:12",
					Commands: []string{"yarn", "install", "--frozen-lockfile", "--non-interactive"},
				},
			},
			Root:        "", // TODO
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}, nil
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
