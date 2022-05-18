package indexing

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

const requeueBackoff = time.Second * 30

// default is false aka index scheduler is enabled
var disableIndexScheduler, _ = strconv.ParseBool(os.Getenv("CODEINTEL_DEPENDENCY_INDEX_SCHEDULER_DISABLED"))

// NewDependencyIndexingScheduler returns a new worker instance that processes
// records from lsif_dependency_indexing_jobs.
func NewDependencyIndexingScheduler(
	dbStore DBStore,
	workerStore dbworkerstore.Store,
	externalServiceStore ExternalServiceStore,
	repoUpdaterClient RepoUpdaterClient,
	gitserverClient GitserverClient,
	enqueuer IndexEnqueuer,
	pollInterval time.Duration,
	numProcessorRoutines int,
	workerMetrics workerutil.WorkerMetrics,
) *workerutil.Worker {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	handler := &dependencyIndexingSchedulerHandler{
		dbStore:       dbStore,
		extsvcStore:   externalServiceStore,
		indexEnqueuer: enqueuer,
		workerStore:   workerStore,
		repoUpdater:   repoUpdaterClient,
		gitserver:     gitserverClient,
	}

	return dbworker.NewWorker(rootContext, workerStore, handler, workerutil.WorkerOptions{
		Name:              "precise_code_intel_dependency_indexing_scheduler_worker",
		NumHandlers:       numProcessorRoutines,
		Interval:          pollInterval,
		Metrics:           workerMetrics,
		HeartbeatInterval: 1 * time.Second,
	})
}

type dependencyIndexingSchedulerHandler struct {
	dbStore       DBStore
	indexEnqueuer IndexEnqueuer
	extsvcStore   ExternalServiceStore
	workerStore   dbworkerstore.Store
	repoUpdater   RepoUpdaterClient
	gitserver     GitserverClient
}

var _ workerutil.Handler = &dependencyIndexingSchedulerHandler{}

// Handle iterates all import monikers associated with a given upload that has
// recently completed processing. Each moniker is interpreted according to its
// scheme to determine the dependent repository and commit. A set of indexing
// jobs are enqueued for each repository and commit pair.
func (h *dependencyIndexingSchedulerHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	if !autoIndexingEnabled() || disableIndexScheduler {
		return nil
	}

	job := record.(dbstore.DependencyIndexingJob)

	if job.ExternalServiceKind != "" {
		externalServices, err := h.extsvcStore.List(ctx, database.ExternalServicesListOptions{
			Kinds: []string{job.ExternalServiceKind},
		})
		if err != nil {
			return errors.Wrap(err, "extsvcStore.List")
		}

		outdatedServices := make(map[int64]time.Duration, len(externalServices))
		for _, externalService := range externalServices {
			if externalService.LastSyncAt.Before(job.ExternalServiceSync) {
				outdatedServices[externalService.ID] = job.ExternalServiceSync.Sub(externalService.LastSyncAt)
			}
		}

		if len(outdatedServices) > 0 {
			if err := h.workerStore.Requeue(ctx, job.ID, time.Now().Add(requeueBackoff)); err != nil {
				return errors.Wrap(err, "store.Requeue")
			}

			entries := make([]log.Field, 0, len(outdatedServices))
			for id, d := range outdatedServices {
				entries = append(entries, log.Duration(fmt.Sprintf("%d", id), d))
			}
			logger.Warn("Requeued dependency indexing job (external services not yet updated)",
				log.Object("outdated_services", entries...))
			return nil
		}
	}

	var errs []error
	scanner, err := h.dbStore.ReferencesForUpload(ctx, job.UploadID)
	if err != nil {
		return errors.Wrap(err, "dbstore.ReferencesForUpload")
	}
	defer func() {
		if closeErr := scanner.Close(); closeErr != nil {
			err = errors.Append(err, errors.Wrap(closeErr, "dbstore.ReferencesForUpload.Close"))
		}
	}()

	repoToPackages := make(map[api.RepoName][]precise.Package)
	var repoNames []api.RepoName
	for {
		packageReference, exists, err := scanner.Next()
		if err != nil {
			return errors.Wrap(err, "dbstore.ReferencesForUpload.Next")
		}
		if !exists {
			break
		}

		pkg := precise.Package{
			Scheme:  packageReference.Package.Scheme,
			Name:    packageReference.Package.Name,
			Version: packageReference.Package.Version,
		}

		repoName, _, ok := enqueuer.InferRepositoryAndRevision(pkg)
		if !ok {
			continue
		}
		repoToPackages[repoName] = append(repoToPackages[repoName], pkg)
		repoNames = append(repoNames, repoName)
	}

	// if this job is not associated with an external service kind that was just synced, then we need to guarantee
	// that the repos are visible to the Sourcegraph instance, else skip them
	if job.ExternalServiceKind == "" {
		for _, repo := range repoNames {
			if _, err := h.repoUpdater.RepoLookup(ctx, repo); errcode.IsNotFound(err) {
				delete(repoToPackages, repo)
			} else if err != nil {
				return errors.Wrapf(err, "repoUpdater.RepoLookup", "repo", repo)
			}
		}
	}

	results, err := h.gitserver.RepoInfo(ctx, repoNames...)
	if err != nil {
		return errors.Wrap(err, "gitserver.RepoInfo")
	}

	for repo, info := range results {
		if !info.Cloned && !info.CloneInProgress { // if the repository doesnt exist
			delete(repoToPackages, repo)
		} else if info.CloneInProgress { // we can't enqueue if still cloning
			return h.workerStore.Requeue(ctx, job.ID, time.Now().Add(requeueBackoff))
		}
	}

	for _, pkgs := range repoToPackages {
		for _, pkg := range pkgs {
			if err := h.indexEnqueuer.QueueIndexesForPackage(ctx, pkg); err != nil {
				errs = append(errs, errors.Wrap(err, "enqueuer.QueueIndexesForPackage"))
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	if len(errs) == 1 {
		return errs[0]
	}

	return errors.Append(nil, errs...)
}
