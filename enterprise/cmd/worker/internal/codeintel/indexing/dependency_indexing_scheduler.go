package indexing

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

var schemeToExternalService = map[string]string{
	"semanticdb": extsvc.KindJVMPackages,
}

// NewDependencyIndexingScheduler returns a new worker instance that processes
// records from lsif_dependency_indexing_jobs.
func NewDependencyIndexingScheduler(
	dbStore DBStore,
	workerStore dbworkerstore.Store,
	externalServiceStore ExternalServiceStore,
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

	var errs []error

	job := record.(dbstore.DependencyIndexingJob)

	shouldIndex, err := h.shouldIndexDependencies(ctx, h.dbStore, job.UploadID)
	if err != nil {
		errs = append(errs, errors.Wrap(err, "indexing.shouldIndexDependencies"))
	}

	scanner, err := h.dbStore.ReferencesForUpload(ctx, job.UploadID)
	if err != nil {
		return errors.Wrap(err, "dbstore.ReferencesForUpload")
	}
	defer func() {
		if closeErr := scanner.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "dbstore.ReferenceIDsAndFilters.Close"))
		}
	}()

	var kinds []string

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

		if shouldIndex {
			if err := h.indexEnqueuer.QueueIndexesForPackage(ctx, pkg); err != nil {
				errs = append(errs, errors.Wrap(err, "enqueuer.QueueIndexesForPackage"))
			}
		}

		extsvcKind, ok := schemeToExternalService[packageReference.Scheme]
		if !ok {
			continue
		}

		if _, err := h.dbStore.InsertCloneableDependencyRepo(ctx, pkg); err != nil {
			errs = append(errs, errors.Wrap(err, "dbstore.InsertCloneableDependencyRepos"))
		}

		if !kindExists(kinds, extsvcKind) {
			kinds = append(kinds, extsvcKind)
		}
	}

	// If len == 0, it will return all external services, which we definitely don't want.
	if len(kinds) > 0 {
		externalServices, err := h.extsvcStore.List(ctx, database.ExternalServicesListOptions{
			Kinds: kinds,
		})
		if err != nil {
			if len(errs) == 0 {
				return errors.Wrap(err, "dbstore.List")
			} else {
				return multierror.Append(err, errs...)
			}
		}

		for _, externalService := range externalServices {
			externalService.NextSyncAt = time.Now()
			err := h.extsvcStore.Upsert(ctx, externalService)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "dbstore.Upsert: error setting next_sync_at for external service %d - %s", externalService.ID, externalService.DisplayName))
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

func kindExists(kinds []string, kind string) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}
