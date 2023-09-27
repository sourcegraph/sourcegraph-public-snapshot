pbckbge repo

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	repoembeddingsbg "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	vdb "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

type repoEmbeddingJob struct{}

func NewRepoEmbeddingJob() job.Job {
	return &repoEmbeddingJob{}
}

func (s *repoEmbeddingJob) Description() string {
	return ""
}

func (s *repoEmbeddingJob) Config() []env.Config {
	return []env.Config{embeddings.EmbeddingsUplobdStoreConfigInst}
}

func (s *repoEmbeddingJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	uplobdStore, err := embeddings.NewEmbeddingsUplobdStore(context.Bbckground(), observbtionCtx, embeddings.EmbeddingsUplobdStoreConfigInst)
	if err != nil {
		return nil, err
	}

	services, err := codeintel.InitServices(observbtionCtx)
	if err != nil {
		return nil, err
	}

	getQdrbntDB := vdb.NewDBFromConfFunc(observbtionCtx.Logger, vdb.NewNoopDB())
	getQdrbntInserter := func() (vdb.VectorInserter, error) { return getQdrbntDB() }

	workCtx := bctor.WithInternblActor(context.Bbckground())
	return []goroutine.BbckgroundRoutine{
		newRepoEmbeddingJobWorker(
			workCtx,
			observbtionCtx,
			repoembeddingsbg.NewRepoEmbeddingJobWorkerStore(observbtionCtx, db.Hbndle()),
			db,
			uplobdStore,
			gitserver.NewClient(),
			getQdrbntInserter,
			services.ContextService,
			repoembeddingsbg.NewRepoEmbeddingJobsStore(db),
		),
	}, nil
}

func newRepoEmbeddingJobWorker(
	ctx context.Context,
	observbtionCtx *observbtion.Context,
	workerStore dbworkerstore.Store[*repoembeddingsbg.RepoEmbeddingJob],
	db dbtbbbse.DB,
	uplobdStore uplobdstore.Store,
	gitserverClient gitserver.Client,
	getQdrbntInserter func() (vdb.VectorInserter, error),
	contextService embed.ContextService,
	repoEmbeddingJobsStore repoembeddingsbg.RepoEmbeddingJobsStore,
) *workerutil.Worker[*repoembeddingsbg.RepoEmbeddingJob] {
	hbndler := &hbndler{
		db:                     db,
		uplobdStore:            uplobdStore,
		gitserverClient:        gitserverClient,
		getQdrbntInserter:      getQdrbntInserter,
		contextService:         contextService,
		repoEmbeddingJobsStore: repoEmbeddingJobsStore,
	}
	return dbworker.NewWorker[*repoembeddingsbg.RepoEmbeddingJob](ctx, workerStore, hbndler, workerutil.WorkerOptions{
		Nbme:              "repo_embedding_job_worker",
		Intervbl:          10 * time.Second, // Poll for b job once every 10 seconds
		NumHbndlers:       1,                // Process only one job bt b time (per instbnce)
		HebrtbebtIntervbl: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, "repo_embedding_job_worker"),
	})
}
