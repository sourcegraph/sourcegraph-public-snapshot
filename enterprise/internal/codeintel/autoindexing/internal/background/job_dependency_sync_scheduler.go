package background

import (
	"context"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewDependencySyncScheduler returns a new worker instance that processes
// records from lsif_dependency_syncing_jobs.
func NewDependencySyncScheduler(
	dependencySyncStore dbworkerstore.Store[shared.DependencySyncingJob],
	uploadSvc UploadService,
	depsSvc DependenciesService,
	store store.Store,
	externalServiceStore ExternalServiceStore,
	metrics workerutil.WorkerObservability,
	pollInterval time.Duration,
) *workerutil.Worker[shared.DependencySyncingJob] {
	rootContext := actor.WithInternalActor(context.Background())
	handler := &dependencySyncSchedulerHandler{
		uploadsSvc:  uploadSvc,
		depsSvc:     depsSvc,
		store:       store,
		workerStore: dependencySyncStore,
		extsvcStore: externalServiceStore,
	}

	return dbworker.NewWorker[shared.DependencySyncingJob](rootContext, dependencySyncStore, handler, workerutil.WorkerOptions{
		Name:              "precise_code_intel_dependency_sync_scheduler_worker",
		Description:       "reads dependency package references from code-intel uploads to be synced to the instance",
		NumHandlers:       1,
		Interval:          pollInterval,
		HeartbeatInterval: 1 * time.Second,
		Metrics:           metrics,
	})
}

type dependencySyncSchedulerHandler struct {
	uploadsSvc  UploadService
	depsSvc     DependenciesService
	store       store.Store
	workerStore dbworkerstore.Store[shared.DependencySyncingJob]
	extsvcStore ExternalServiceStore
}

// For mocking in tests
var autoIndexingEnabled = conf.CodeIntelAutoIndexingEnabled

var schemeToExternalService = map[string]string{
	dependencies.JVMPackagesScheme:    extsvc.KindJVMPackages,
	dependencies.NpmPackagesScheme:    extsvc.KindNpmPackages,
	dependencies.PythonPackagesScheme: extsvc.KindPythonPackages,
	dependencies.RustPackagesScheme:   extsvc.KindRustPackages,
	dependencies.RubyPackagesScheme:   extsvc.KindRubyPackages,
}

func (h *dependencySyncSchedulerHandler) Handle(ctx context.Context, logger log.Logger, job shared.DependencySyncingJob) error {
	if !autoIndexingEnabled() {
		return nil
	}

	scanner, err := h.uploadsSvc.ReferencesForUpload(ctx, job.UploadID)
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

		pkgRef, err := newPackage(packageReference.Package)
		if err != nil {
			// Indexers can potentially create package references with bad names,
			// which are no longer recognized by the package manager. In such a
			// case, it doesn't make sense to add a bad package as a dependency repo.
			logger.Warn("package referenced by upload was invalid",
				log.Error(err),
				log.String("name", packageReference.Name),
				log.String("version", packageReference.Version),
				log.Int("dumpId", packageReference.DumpID))
			continue
		}
		pkg := *pkgRef

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

	shouldIndex, err := h.shouldIndexDependencies(ctx, h.uploadsSvc, job.UploadID)
	if err != nil {
		return err
	}

	if shouldIndex {
		// If we saw a kind that's not in schemeToExternalService, then kinds contains an empty string key
		for kind := range kinds {
			if _, err := h.store.InsertDependencyIndexingJob(ctx, job.UploadID, kind, nextSync); err != nil {
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
func newPackage(pkg uploadsshared.Package) (*precise.Package, error) {
	p := precise.Package{
		Scheme:  pkg.Scheme,
		Manager: pkg.Manager,
		Name:    pkg.Name,
		Version: pkg.Version,
	}

	switch pkg.Scheme {
	case dependencies.JVMPackagesScheme:
		p.Name = strings.TrimPrefix(p.Name, "maven/")
		p.Name = strings.ReplaceAll(p.Name, "/", ":")
	case dependencies.NpmPackagesScheme:
		if _, err := reposource.ParseNpmPackageFromPackageSyntax(reposource.PackageName(p.Name)); err != nil {
			return nil, err
		}
	case "scip-python":
		// Override scip-python scheme so that we are able to autoindex
		// index.scip created by scip-python
		p.Scheme = dependencies.PythonPackagesScheme
	}

	return &p, nil
}

func (h *dependencySyncSchedulerHandler) insertDependencyRepo(ctx context.Context, pkg precise.Package) (new bool, err error) {
	inserted, err := h.depsSvc.UpsertDependencyRepos(ctx, []dependencies.Repo{
		{
			Name:    reposource.PackageName(pkg.Name),
			Scheme:  pkg.Scheme,
			Version: pkg.Version,
		},
	})
	if err != nil {
		return false, errors.Wrap(err, "dbstore.InsertCloneableDependencyRepos")
	}
	return len(inserted) != 0, nil
}

// shouldIndexDependencies returns true if the given upload should undergo dependency
// indexing. Currently, we're only enabling dependency indexing for a repositories that
// were indexed via lsif-go, scip-java, lsif-tsc and scip-typescript.
func (h *dependencySyncSchedulerHandler) shouldIndexDependencies(ctx context.Context, store UploadService, uploadID int) (bool, error) {
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
		upload.Indexer == "scip-python" ||
		upload.Indexer == "scip-ruby" ||
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
