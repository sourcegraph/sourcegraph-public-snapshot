package worker

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TODO(sasha) rename to `repo_update_worker` file and struct
func MakeRepoCloneWorker(ctx context.Context, observationCtx *observation.Context, db database.DB, gitserver *server.Server, hostname string) *workerutil.Worker[*types.RepoCloneJob] {
	numHandlers := 1
	store := MakeStore(observationCtx, db.Handle())
	handler := MakeRepoCloneHandler(observationCtx, database.RepoCloneJobStoreWith(db), hostname, gitserver)
	return dbworker.NewWorker[*types.RepoCloneJob](ctx, store, handler, workerutil.WorkerOptions{
		Name:              "repo_clone_worker",
		Interval:          time.Second,
		HeartbeatInterval: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "repo_clone_worker"),
		NumHandlers:       numHandlers,
	})
}

func MakeStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[*types.RepoCloneJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[*types.RepoCloneJob]{
		Name:              "repo_clone_worker_store",
		TableName:         "repo_clone_jobs",
		ColumnExpressions: database.RepoCloneJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(database.ScanRepoCloneJob),
		// TODO(sasha) keep backoff in mind
		OrderByExpression: sqlf.Sprintf("clone, id"),
		MaxNumResets:      5,
		StalledMaxAge:     time.Second * 30,
	})
}

func MakeRepoCloneHandler(observationCtx *observation.Context, jobsStore database.RepoCloneJobStore, hostname string, gitserver *server.Server) *repoCloneHandler {
	logger := observationCtx.Logger.Scoped("RepoCloneWorker", "Repository cloning worker")
	return &repoCloneHandler{
		logger:    logger,
		jobsStore: jobsStore,
		hostname:  hostname,
		gitserver: gitserver,
	}
}

type repoCloneHandler struct {
	logger    log.Logger
	jobsStore database.RepoCloneJobStore
	hostname  string
	gitserver *server.Server
}

func (h *repoCloneHandler) Handle(ctx context.Context, logger log.Logger, record *types.RepoCloneJob) error {
	h.logger.Info(
		"$$$$$$$$$$$$$$$$$ Handling repo clone job $$$$$$$$$$$$$$$$$$$$",
		log.String("repo_name", record.RepoName),
		log.String("job_gitserver_address", record.GitserverAddress),
		log.String("worker_gitserver_address", h.hostname),
	)
	// TODO(sasha): check the addrForRepo again and whether it matches with h.hostname.
	// TODO(sasha): if the address doesn't match, we can update the job with a new address and enqueue it again.
	// TODO(sasha) check string <-> api.RepoName back and forth
	resp := h.gitserver.HandleRepoUpdateRequest(ctx, &protocol.RepoUpdateRequest{Repo: api.RepoName(record.RepoName), Since: time.Duration(record.UpdateAfter) * time.Second}, logger)
	if resp.Error != "" {
		// TODO(sasha): error handling: if the repo is not found, return errcode.MakeNonRetryable(err)
		return errors.New(resp.Error)
	}
	return nil
}

// PreDequeue adds a predicate to only fetch jobs which should be processed by
// this exact gitserver instance.
func (h *repoCloneHandler) PreDequeue(_ context.Context, _ log.Logger) (bool, any, error) {
	return true, []*sqlf.Query{sqlf.Sprintf("gitserver_address = %s", h.hostname)}, nil
}
