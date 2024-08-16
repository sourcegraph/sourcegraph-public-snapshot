package dependencies

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/cast"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewDependencyIndexingScheduler returns a new worker instance that processes
// records from lsif_dependency_indexing_jobs.
func NewDependencyIndexingScheduler(
	dependencyIndexingStore dbworkerstore.Store[dependencyIndexingJob],
	uploadSvc UploadService,
	repoStore ReposStore,
	externalServiceStore ExternalServiceStore,
	gitserverRepoStore GitserverRepoStore,
	indexEnqueuer IndexEnqueuer,
	metrics workerutil.WorkerObservability,
	config *Config,
) *workerutil.Worker[dependencyIndexingJob] {
	rootContext := actor.WithInternalActor(context.Background())

	handler := &dependencyIndexingSchedulerHandler{
		uploadsSvc:         uploadSvc,
		repoStore:          repoStore,
		extsvcStore:        externalServiceStore,
		gitserverRepoStore: gitserverRepoStore,
		indexEnqueuer:      indexEnqueuer,
		workerStore:        dependencyIndexingStore,
	}

	return dbworker.NewWorker[dependencyIndexingJob](rootContext, dependencyIndexingStore, handler, workerutil.WorkerOptions{
		Name:              "precise_code_intel_dependency_indexing_scheduler_worker",
		Description:       "queues code-intel auto-indexing jobs for dependency packages",
		NumHandlers:       config.DependencyIndexerSchedulerConcurrency,
		Interval:          config.DependencyIndexerSchedulerPollInterval,
		Metrics:           metrics,
		HeartbeatInterval: 1 * time.Second,
	})
}

type dependencyIndexingSchedulerHandler struct {
	uploadsSvc         UploadService
	repoStore          ReposStore
	indexEnqueuer      IndexEnqueuer
	extsvcStore        ExternalServiceStore
	gitserverRepoStore GitserverRepoStore
	workerStore        dbworkerstore.Store[dependencyIndexingJob]
}

const requeueBackoff = time.Second * 30

// default is false aka index scheduler is enabled
var disableIndexScheduler, _ = strconv.ParseBool(os.Getenv("CODEINTEL_DEPENDENCY_INDEX_SCHEDULER_DISABLED"))

var _ workerutil.Handler[dependencyIndexingJob] = &dependencyIndexingSchedulerHandler{}

// Handle iterates all import monikers associated with a given upload that has
// recently completed processing. Each moniker is interpreted according to its
// scheme to determine the dependent repository and commit. A set of indexing
// jobs are enqueued for each repository and commit pair.
func (h *dependencyIndexingSchedulerHandler) Handle(ctx context.Context, logger log.Logger, job dependencyIndexingJob) error {
	if !autoIndexingEnabled() || disableIndexScheduler {
		return nil
	}

	if job.ExternalServiceKind != "" {
		externalServices, err := h.extsvcStore.List(ctx, database.ExternalServicesListOptions{
			Kinds: []string{job.ExternalServiceKind},
		})
		if err != nil {
			return errors.Wrap(err, "extsvcStore.List")
		}

		if len(externalServices) == 0 {
			logger.Info("no external services returned, auto-index jobs cannot be added for package repos of the given external service type if it does not exist",
				log.String("extsvcKind", job.ExternalServiceKind))
			return nil
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

	scanner, err := h.uploadsSvc.ReferencesForUpload(ctx, job.UploadID)
	if err != nil {
		return errors.Wrap(err, "dbstore.ReferencesForUpload")
	}
	defer func() {
		if closeErr := scanner.Close(); closeErr != nil {
			err = errors.Append(err, errors.Wrap(closeErr, "dbstore.ReferencesForUpload.Close"))
		}
	}()

	repoToPackages := make(map[api.RepoName][]dependencies.MinimialVersionedPackageRepo)
	var repoNames []api.RepoName
	for {
		packageReference, exists, err := scanner.Next()
		if err != nil {
			return errors.Wrap(err, "dbstore.ReferencesForUpload.Next")
		}
		if !exists {
			break
		}

		pkg := dependencies.MinimialVersionedPackageRepo{
			Scheme:  packageReference.Scheme,
			Name:    reposource.PackageName(packageReference.Name),
			Version: packageReference.Version,
		}

		repoName, _, ok := inference.InferRepositoryAndRevision(pkg)
		if !ok {
			continue
		}
		repoToPackages[repoName] = append(repoToPackages[repoName], pkg)
		repoNames = append(repoNames, repoName)
	}

	// No dependencies found, we can return early.
	if len(repoNames) == 0 {
		return nil
	}

	// if this job is not associated with an external service kind that was just synced, then we need to guarantee
	// that the repos are visible to the Sourcegraph instance, else skip them
	if job.ExternalServiceKind == "" {
		repoNameStrings := cast.ToStrings(repoNames)
		sort.Strings(repoNameStrings)

		listedRepos, err := h.repoStore.ListMinimalRepos(ctx, database.ReposListOptions{
			Names:   repoNameStrings,
			OrderBy: []database.RepoListSort{{Field: database.RepoListName}},
		})
		if err != nil {
			logger.Error("error listing repositories, continuing", log.Error(err), log.Int("numRepos", len(repoNameStrings)))
		}

		listedRepoNames := make([]api.RepoName, 0, len(listedRepos))
		for _, repo := range listedRepos {
			listedRepoNames = append(listedRepoNames, repo.Name)
		}

		// for any repos that are not known to the instance, we need to sync them if on dot-com,
		// otherwise skip them.
		difference := setDifference(repoNames, listedRepoNames)

		for _, repo := range difference {
			delete(repoToPackages, repo)
		}
	}

	results, err := h.gitserverRepoStore.GetByNames(ctx, repoNames...)
	if err != nil {
		return errors.Wrap(err, "gitserver.RepoInfo")
	}

	for _, repoName := range repoNames {
		repoInfo, ok := results[repoName]
		if !ok || repoInfo.CloneStatus != types.CloneStatusCloned && repoInfo.CloneStatus != types.CloneStatusCloning {
			delete(repoToPackages, repoName)
		} else if repoInfo.CloneStatus == types.CloneStatusCloning { // we can't enqueue if still cloning
			return h.workerStore.Requeue(ctx, job.ID, time.Now().Add(requeueBackoff))
		}
	}

	var errs []error
	for _, pkgs := range repoToPackages {
		for _, pkg := range pkgs {
			if err := h.indexEnqueuer.QueueAutoIndexJobsForPackage(ctx, pkg); err != nil {
				errs = append(errs, errors.Wrap(err, "enqueuer.QueueAutoIndexJobsForPackage"))
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

// Returns the set of elements in superset that are not in subset
// invariants:
//   - superset is, of course, a superset of subset.
//   - subset does not contain duplicates
func setDifference[T comparable](superset, subset []T) (ret []T) {
	j := 0
	for i, val := range superset {
		if i > 0 && val == superset[i-1] {
			continue
		}
		if j > len(subset)-1 {
			ret = append(ret, val)
			continue
		}

		if val == subset[j] {
			j++
		} else if val != subset[j] {
			ret = append(ret, val)
		}
	}

	return
}
