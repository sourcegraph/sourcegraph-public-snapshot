package worker

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func MakeRepoUpdateWorker(ctx context.Context, observationCtx *observation.Context, db database.DB, gitserver *server.Server, hostname string) *workerutil.Worker[types.RepoUpdateJob] {
	numHandlers := conf.GitMaxConcurrentClones()
	store := MakeStore(observationCtx, db.Handle())
	handler := MakeRepoCloneHandler(observationCtx, db, database.RepoUpdateJobStoreWith(db), hostname, gitserver)
	conf.GitLongCommandTimeout()
	return dbworker.NewWorker[types.RepoUpdateJob](ctx, store, handler, workerutil.WorkerOptions{
		Name:                 "repo_clone_worker",
		Interval:             time.Second,
		HeartbeatInterval:    10 * time.Second,
		Metrics:              workerutil.NewMetrics(observationCtx, "repo_clone_worker"),
		NumHandlers:          numHandlers,
		MaximumRuntimePerJob: conf.GitLongCommandTimeout(),
	})
}

func MakeStore(observationCtx *observation.Context, dbHandle basestore.TransactableHandle) dbworkerstore.Store[types.RepoUpdateJob] {
	return dbworkerstore.New(observationCtx, dbHandle, dbworkerstore.Options[types.RepoUpdateJob]{
		Name:      "repo_clone_worker_store",
		TableName: "repo_update_jobs",
		// We alias the view with a table name to simplify columns names handling.
		ViewName:          "repo_update_jobs_with_repo_name repo_update_jobs",
		ColumnExpressions: database.FullRepoUpdateJobColumns,
		Scan:              dbworkerstore.BuildWorkerScan(database.ScanFullRepoUpdateJob),
		OrderByExpression: sqlf.Sprintf("repo_update_jobs.priority DESC, repo_update_jobs.process_after ASC NULLS FIRST, repo_update_jobs.id ASC"),
		MaxNumResets:      1,
		// In case of gitserver restart we don't want to keep processing jobs which will never end for long.
		StalledMaxAge: 20 * time.Second,
	})
}

func MakeRepoCloneHandler(observationCtx *observation.Context, db database.DB, jobsStore database.RepoUpdateJobStore, hostname string, gitserver *server.Server) *repoCloneHandler {
	logger := observationCtx.Logger.Scoped("RepoCloneWorker", "Repository cloning worker")
	return &repoCloneHandler{
		logger:    logger,
		db:        db,
		jobsStore: jobsStore,
		hostname:  hostname,
		gitserver: gitserver,
	}
}

type repoCloneHandler struct {
	logger    log.Logger
	db        database.DB
	jobsStore database.RepoUpdateJobStore
	hostname  string
	// Keep this dependency until finished the project and maybe after that we can
	// remove this completely.
	gitserver *server.Server
}

func (h *repoCloneHandler) Handle(ctx context.Context, logger log.Logger, record types.RepoUpdateJob) error {
	repoName := record.RepositoryName
	resp, err := h.gitserver.HandleRepoUpdateRequest(ctx, &protocol.RepoUpdateRequest{Repo: repoName}, logger)
	if err != nil {
		if errcode.IsNotFound(err) {
			return errcode.MakeNonRetryable(err)
		}
		return errors.New(resp.Error)
	}
	if resp.LastFetched != nil && resp.LastChanged != nil {
		if err := h.jobsStore.SaveUpdateJobResults(ctx,
			record.ID,
			database.SaveUpdateJobResultsOpts{LastFetched: *resp.LastFetched, LastChanged: *resp.LastChanged},
		); err != nil {
			return err
		}
	}
	return nil
}

// PreDequeue adds a predicate to only fetch jobs which should be processed by
// this exact gitserver instance.
func (h *repoCloneHandler) PreDequeue(_ context.Context, _ log.Logger) (bool, any, error) {
	addresses := gitserver.NewGitserverAddresses(h.db, conf.Get()).Addresses
	addrPredicates := make([]*sqlf.Query, 0, len(addresses))
	for _, addr := range addresses {
		addrPredicates = append(addrPredicates, sqlf.Sprintf("%s", addr))
	}
	return true, []*sqlf.Query{sqlf.Sprintf("addr_for_repo(repo_update_jobs.repository_name, ARRAY[%s]) = %s", sqlf.Join(addrPredicates, ","), h.hostname)}, nil
}
