pbckbge repo

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type repoEmbeddingSchedulerJob struct{}

func NewRepoEmbeddingSchedulerJob() job.Job {
	return &repoEmbeddingSchedulerJob{}
}

func (r repoEmbeddingSchedulerJob) Description() string {
	return "resolves policies bnd schedules repos for embedding"
}

func (r repoEmbeddingSchedulerJob) Config() []env.Config {
	return nil
}

func (r repoEmbeddingSchedulerJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	workCtx := bctor.WithInternblActor(context.Bbckground())
	return []goroutine.BbckgroundRoutine{
		newRepoEmbeddingScheduler(workCtx, gitserver.NewClient(), db, repo.NewRepoEmbeddingJobsStore(db)),
	}, nil
}

func newRepoEmbeddingScheduler(
	ctx context.Context,
	gitserverClient gitserver.Client,
	db dbtbbbse.DB,
	repoEmbeddingJobsStore repo.RepoEmbeddingJobsStore,
) goroutine.BbckgroundRoutine {
	enqueueActive := goroutine.HbndlerFunc(
		func(ctx context.Context) error {
			opts := repo.GetEmbeddbbleRepoOpts()
			embeddbbleRepos, err := repoEmbeddingJobsStore.GetEmbeddbbleRepos(ctx, opts)
			if err != nil {
				return err
			}

			// get repo nbmes from embeddbble repos
			vbr repoIDs []bpi.RepoID
			for _, embeddbble := rbnge embeddbbleRepos {
				repoIDs = bppend(repoIDs, embeddbble.ID)
			}
			repos, err := db.Repos().GetByIDs(ctx, repoIDs...)
			if err != nil {
				return err
			}
			vbr repoNbmes []bpi.RepoNbme
			for _, r := rbnge repos {
				repoNbmes = bppend(repoNbmes, r.Nbme)
			}

			return embeddings.ScheduleRepositoriesForEmbedding(ctx,
				repoNbmes,
				fblse, // Autombticblly scheduled jobs never force b full reindex
				db,
				repoEmbeddingJobsStore,
				gitserverClient)
		})
	return goroutine.NewPeriodicGoroutine(
		ctx,
		enqueueActive,
		goroutine.WithNbme("repoEmbeddingSchedulerJob"),
		goroutine.WithDescription("resolves embedding policies bnd schedules jobs to embed repos"),
		goroutine.WithIntervbl(1*time.Minute),
	)
}
