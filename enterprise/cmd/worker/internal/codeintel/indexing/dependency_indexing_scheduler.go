package indexing

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/enqueuer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// NewDependencyIndexingScheduler returns a new worker instance that processes
// records from lsif_dependency_indexing_queueing_jobs.
func NewDependencyIndexingScheduler(
	dbStore DBStore,
	workerStore dbworkerstore.Store,
	externalServiceStore ExternalServiceStore,
	gitserver GitserverClient,
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
	gitserver     GitserverClient
}

var _ workerutil.Handler = &dependencyIndexingSchedulerHandler{}

// Handle iterates all import monikers associated with a given upload that has
// recently completed processing. Each moniker is interpreted according to its
// scheme to determine the dependent repository and commit. A set of indexing
// jobs are enqueued for each repository and commit pair.
func (h *dependencyIndexingSchedulerHandler) Handle(ctx context.Context, record workerutil.Record) error {
	if !indexSchedulerEnabled() {
		return nil
	}

	job := record.(dbstore.DependencyIndexingQueueingJob)

	shouldIndex, err := h.shouldIndexDependencies(ctx, h.dbStore, job.UploadID)
	if err != nil {
		return errors.Wrap(err, "indexing.shouldIndexDependencies")
	}
	if !shouldIndex {
		return nil
	}

	if job.ExternalServiceKind != "" {
		externalServices, err := h.extsvcStore.List(ctx, database.ExternalServicesListOptions{
			Kinds: []string{job.ExternalServiceKind},
		})
		if err != nil {
			return errors.Wrap(err, "dbstore.List")
		}

		for _, externalService := range externalServices {
			if externalService.LastSyncAt.Before(job.ExternalServiceSync) {
				return h.workerStore.Requeue(ctx, job.ID, time.Now().Add(time.Second*10))
			}
		}
	}

	var errs []error
	scanner, err := h.dbStore.ReferencesForUpload(ctx, job.UploadID)
	if err != nil {
		return errors.Wrap(err, "dbstore.ReferencesForUpload")
	}
	defer func() {
		if closeErr := scanner.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "dbstore.ReferencesForUpload.Close"))
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

		name, _, ok := enqueuer.InferRepositoryAndRevision(pkg)
		if !ok {
			continue
		}
		repoToPackages[api.RepoName(name)] = append(repoToPackages[api.RepoName(name)], pkg)
		repoNames = append(repoNames, api.RepoName(name))
	}

	results, err := h.gitserver.RepoInfo(ctx, repoNames...)
	if err != nil {
		return err
	}

	for repo, info := range results {
		if !info.Cloned && !info.CloneInProgress { // if the repository doesnt exist
			delete(repoToPackages, repo)
		} else if info.CloneInProgress { // we can't enqueue if still cloning
			return h.workerStore.Requeue(ctx, job.ID, time.Now().Add(time.Second*10))
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

	return multierror.Append(nil, errs...)
}

// shouldIndexDependencies returns true if the given upload should undergo dependency
// indexing. Currently, we're only enabling dependency indexing for a repositories that
// were indexed via lsif-go and lsif-java.
func (h *dependencyIndexingSchedulerHandler) shouldIndexDependencies(ctx context.Context, store DBStore, uploadID int) (bool, error) {
	upload, _, err := store.GetUploadByID(ctx, uploadID)
	if err != nil {
		return false, errors.Wrap(err, "dbstore.GetUploadByID")
	}

	return upload.Indexer == "lsif-go" || upload.Indexer == "lsif-java", nil
}
