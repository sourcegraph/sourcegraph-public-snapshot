package indexing

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

var schemeToExternalService = map[string]string{
	dependencies.JVMPackagesScheme:  extsvc.KindJVMPackages,
	dependencies.NpmPackagesScheme:  extsvc.KindNpmPackages,
	dependencies.RustPackagesScheme: extsvc.KindRustPackages,
}

// NewDependencySyncScheduler returns a new worker instance that processes
// records from lsif_dependency_syncing_jobs.
func NewDependencySyncScheduler(
	dbStore DBStore,
	workerStore dbworkerstore.Store,
	externalServiceStore ExternalServiceStore,
	metrics workerutil.WorkerMetrics,
	observationContext *observation.Context,
) *workerutil.Worker {
	// Init metrics here now after we've moved the autoindexing scheduler
	// into the autoindexing service
	newOperations(observationContext)

	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	handler := &dependencySyncSchedulerHandler{
		dbStore:     dbStore,
		workerStore: workerStore,
		extsvcStore: externalServiceStore,
	}

	return dbworker.NewWorker(rootContext, workerStore, handler, workerutil.WorkerOptions{
		Name:              "precise_code_intel_dependency_sync_scheduler_worker",
		NumHandlers:       1,
		Interval:          time.Second * 5,
		HeartbeatInterval: 1 * time.Second,
		Metrics:           metrics,
	})
}

type dependencySyncSchedulerHandler struct {
	dbStore     DBStore
	workerStore dbworkerstore.Store
	extsvcStore ExternalServiceStore
}

func (h *dependencySyncSchedulerHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) error {
	if !autoIndexingEnabled() {
		return nil
	}

	job := record.(dbstore.DependencySyncingJob)

	scanner, err := h.dbStore.ReferencesForUpload(ctx, job.UploadID)
	if err != nil {
		return errors.Wrap(err, "dbstore.ReferencesForUpload")
	}
	defer func() {
		if closeErr := scanner.Close(); closeErr != nil {
			err = errors.Append(err, errors.Wrap(closeErr, "dbstore.ReferencesForUpload.Close"))
		}
	}()

	var (
		kinds                      = map[string]struct{}{}
		oldDependencyReposInserted int
		newDependencyReposInserted int
		errs                       []error
	)

	for {
		packageReference, exists, err := scanner.Next()
		if err != nil {
			return errors.Wrap(err, "dbstore.ReferencesForUpload.Next")
		}
		if !exists {
			break
		}

		pkg := newPackage(packageReference.Package)

		extsvcKind, ok := schemeToExternalService[pkg.Scheme]
		// add entry for empty string/kind here so dependencies such as lsif-go ones still get
		// an associated dependency indexing job
		kinds[extsvcKind] = struct{}{}
		if !ok {
			continue
		}

		new, err := h.insertDependencyRepo(ctx, pkg)
		if err != nil {
			errs = append(errs, err)
		} else if new {
			newDependencyReposInserted++
		} else {
			oldDependencyReposInserted++
		}
	}

	var nextSync time.Time
	kindsArray := kindsToArray(kinds)
	// If len == 0, it will return all external services, which we definitely don't want.
	if len(kindsArray) > 0 {
		nextSync = time.Now()
		externalServices, err := h.extsvcStore.List(ctx, database.ExternalServicesListOptions{
			Kinds: kindsArray,
		})
		if err != nil {
			if len(errs) == 0 {
				return errors.Wrap(err, "dbstore.List")
			} else {
				return errors.Append(err, errs...)
			}
		}

		logger.Info("syncing external services",
			log.Int("upload", job.UploadID),
			log.Int("numExtSvc", len(externalServices)),
			log.Strings("schemaKinds", kindsArray),
			log.Int("newRepos", newDependencyReposInserted),
			log.Int("existingInserts", oldDependencyReposInserted))

		for _, externalService := range externalServices {
			externalService.NextSyncAt = nextSync
			err := h.extsvcStore.Upsert(ctx, externalService)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "extsvcStore.Upsert: error setting next_sync_at for external service %d - %s", externalService.ID, externalService.DisplayName))
			}
		}
	} else {
		logger.Info("no package schema kinds to sync external services for", log.Int("upload", job.UploadID), log.Int("job", job.ID))
	}

	shouldIndex, err := h.shouldIndexDependencies(ctx, h.dbStore, job.UploadID)
	if err != nil {
		return err
	}

	if shouldIndex {
		// If we saw a kind that's not in schemeToExternalService, then kinds contains an empty string key
		for kind := range kinds {
			if _, err := h.dbStore.InsertDependencyIndexingJob(ctx, job.UploadID, kind, nextSync); err != nil {
				errs = append(errs, errors.Wrap(err, "dbstore.InsertDependencyIndexingJob"))
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

// newPackage constructs a precise.Package from the given shared.Package,
// applying any normalization or necessary transformations that lsif uploads
// require for internal consistency.
func newPackage(pkg shared.Package) precise.Package {
	p := precise.Package{
		Scheme:  pkg.Scheme,
		Name:    pkg.Name,
		Version: pkg.Version,
	}

	switch pkg.Scheme {
	case dependencies.JVMPackagesScheme:
		p.Name = strings.TrimPrefix(p.Name, "maven/")
		p.Name = strings.ReplaceAll(p.Name, "/", ":")
	}

	return p
}

func (h *dependencySyncSchedulerHandler) insertDependencyRepo(ctx context.Context, pkg precise.Package) (new bool, err error) {
	ctx, _, endObservation := dependencyReposOps.InsertCloneableDependencyRepo.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{pkg.Scheme},
	})
	defer func() {
		endObservation(1, observation.Args{MetricLabelValues: []string{strconv.FormatBool(new)}})
	}()

	new, err = h.dbStore.InsertCloneableDependencyRepo(ctx, pkg)
	if err != nil {
		return new, errors.Wrap(err, "dbstore.InsertCloneableDependencyRepos")
	}
	return new, nil
}

// shouldIndexDependencies returns true if the given upload should undergo dependency
// indexing. Currently, we're only enabling dependency indexing for a repositories that
// were indexed via lsif-go, scip-java, lsif-tsc and scip-typescript.
func (h *dependencySyncSchedulerHandler) shouldIndexDependencies(ctx context.Context, store DBStore, uploadID int) (bool, error) {
	upload, _, err := store.GetUploadByID(ctx, uploadID)
	if err != nil {
		return false, errors.Wrap(err, "dbstore.GetUploadByID")
	}

	return upload.Indexer == "lsif-go" ||
		upload.Indexer == "scip-java" ||
		upload.Indexer == "lsif-java" ||
		upload.Indexer == "lsif-tsc" ||
		upload.Indexer == "scip-typescript" ||
		upload.Indexer == "lsif-typescript" ||
		upload.Indexer == "rust-analyzer", nil
}

func kindsToArray(k map[string]struct{}) (s []string) {
	for kind := range k {
		if kind != "" {
			s = append(s, kind)
		}
	}
	return
}
