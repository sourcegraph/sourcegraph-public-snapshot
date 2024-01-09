package repo

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type repoEmbeddingSchedulerJob struct{}

func NewRepoEmbeddingSchedulerJob() job.Job {
	return &repoEmbeddingSchedulerJob{}
}

func (r repoEmbeddingSchedulerJob) Description() string {
	return "resolves policies and schedules repos for embedding"
}

func (r repoEmbeddingSchedulerJob) Config() []env.Config {
	return nil
}

func (r repoEmbeddingSchedulerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	workCtx := actor.WithInternalActor(context.Background())
	return []goroutine.BackgroundRoutine{
		newRepoEmbeddingScheduler(workCtx, gitserver.NewClient(), db, repo.NewRepoEmbeddingJobsStore(db)),
	}, nil
}

func newRepoEmbeddingScheduler(
	ctx context.Context,
	gitserverClient gitserver.Client,
	db database.DB,
	repoEmbeddingJobsStore repo.RepoEmbeddingJobsStore,
) goroutine.BackgroundRoutine {
	enqueueActive := goroutine.HandlerFunc(
		func(ctx context.Context) error {
			opts := repo.GetEmbeddableRepoOpts()
			embeddableRepos, err := repoEmbeddingJobsStore.GetEmbeddableRepos(ctx, opts)
			if err != nil {
				return err
			}

			var repoIDs []api.RepoID
			for _, embeddable := range embeddableRepos {
				repoIDs = append(repoIDs, embeddable.ID)
			}

			return embeddings.ScheduleRepositoriesForPolicy(ctx,
				repoIDs,
				db,
				repoEmbeddingJobsStore,
				gitserverClient)
		})
	return goroutine.NewPeriodicGoroutine(
		ctx,
		enqueueActive,
		goroutine.WithName("repoEmbeddingSchedulerJob"),
		goroutine.WithDescription("resolves embedding policies and schedules jobs to embed repos"),
		goroutine.WithInterval(15*time.Minute),
	)
}
