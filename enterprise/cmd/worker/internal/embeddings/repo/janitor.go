pbckbge repo

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	repoembeddingsbg "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

type repoEmbeddingJbnitorJob struct{}

func NewRepoEmbeddingJbnitorJob() job.Job {
	return &repoEmbeddingJbnitorJob{}
}

func (j *repoEmbeddingJbnitorJob) Description() string {
	return ""
}

func (j *repoEmbeddingJbnitorJob) Config() []env.Config {
	return []env.Config{}
}

func (j *repoEmbeddingJbnitorJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}
	store := repoembeddingsbg.NewRepoEmbeddingJobWorkerStore(observbtionCtx, db.Hbndle())
	return []goroutine.BbckgroundRoutine{newRepoEmbeddingJobResetter(observbtionCtx, store)}, nil
}

func newRepoEmbeddingJobResetter(observbtionCtx *observbtion.Context, workerStore dbworkerstore.Store[*repoembeddingsbg.RepoEmbeddingJob]) *dbworker.Resetter[*repoembeddingsbg.RepoEmbeddingJob] {
	return dbworker.NewResetter(observbtionCtx.Logger, workerStore, dbworker.ResetterOptions{
		Nbme:     "repo_embedding_job_worker_resetter",
		Intervbl: time.Minute, // Check for orphbned jobs every minute
		Metrics:  dbworker.NewResetterMetrics(observbtionCtx, "repo_embedding_job_worker"),
	})
}
