package background

import (
	"context"
	"os"
	"time"

	"github.com/derision-test/glock"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type BackgroundJob interface {
	NewDependencyIndexingScheduler(pollInterval time.Duration, numHandlers int) *workerutil.Worker
	NewDependencySyncScheduler(pollInterval time.Duration) *workerutil.Worker
	NewDependencyIndexResetter(interval time.Duration) *dbworker.Resetter
	NewIndexResetter(interval time.Duration) *dbworker.Resetter
	NewOnDemandScheduler(interval time.Duration, batchSize int) goroutine.BackgroundRoutine
	NewScheduler(interval time.Duration, repositoryProcessDelay time.Duration, repositoryBatchSize int, policyBatchSize int) goroutine.BackgroundRoutine
	NewJanitor(
		interval time.Duration,
		minimumTimeSinceLastCheck time.Duration,
		commitResolverBatchSize int,
		commitResolverMaximumCommitLag time.Duration,
	) goroutine.BackgroundRoutine

	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []types.Index, err error)
	QueueIndexesForPackage(ctx context.Context, pkg precise.Package) (err error)
	InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error)

	InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]config.IndexJob, error)
	InferIndexJobHintsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error)

	WorkerutilStore() dbworkerstore.Store
	DependencySyncStore() dbworkerstore.Store
	DependencyIndexingStore() dbworkerstore.Store
}

type backgroundJob struct {
	uploadSvc    shared.UploadService
	depsSvc      DependenciesService
	inferenceSvc shared.InferenceService
	policiesSvc  PoliciesService

	policyMatcher   PolicyMatcher
	repoUpdater     shared.RepoUpdaterClient
	gitserverClient shared.GitserverClient

	store                   store.Store
	repoStore               ReposStore
	workerutilStore         dbworkerstore.Store
	gitserverRepoStore      GitserverRepoStore
	dependencySyncStore     dbworkerstore.Store
	externalServiceStore    ExternalServiceStore
	dependencyIndexingStore dbworkerstore.Store

	operations *operations
	clock      glock.Clock
	logger     log.Logger

	metrics                *resetterMetrics
	janitorMetrics         *janitorMetrics
	depencencySyncMetrics  workerutil.WorkerMetrics
	depencencyIndexMetrics workerutil.WorkerMetrics
}

func New(
	db database.DB,
	store store.Store,
	uploadSvc shared.UploadService,
	depsSvc DependenciesService,
	policiesSvc PoliciesService,
	inferenceSvc shared.InferenceService,
	policyMatcher PolicyMatcher,
	gitserverClient shared.GitserverClient,
	repoUpdater shared.RepoUpdaterClient,
	observationContext *observation.Context,
) BackgroundJob {
	repoStore := db.Repos()
	gitserverRepoStore := db.GitserverRepos()
	externalServiceStore := db.ExternalServices()
	workerutilStore := dbworkerstore.NewWithMetrics(db.Handle(), indexWorkerStoreOptions, observationContext)
	dependencySyncStore := dbworkerstore.NewWithMetrics(db.Handle(), dependencySyncingJobWorkerStoreOptions, observationContext)
	dependencyIndexingStore := dbworkerstore.NewWithMetrics(db.Handle(), dependencyIndexingJobWorkerStoreOptions, observationContext)

	return &backgroundJob{
		uploadSvc:    uploadSvc,
		depsSvc:      depsSvc,
		inferenceSvc: inferenceSvc,
		policiesSvc:  policiesSvc,

		policyMatcher:   policyMatcher,
		repoUpdater:     repoUpdater,
		gitserverClient: gitserverClient,

		store:                   store,
		repoStore:               repoStore,
		workerutilStore:         workerutilStore,
		gitserverRepoStore:      gitserverRepoStore,
		dependencySyncStore:     dependencySyncStore,
		externalServiceStore:    externalServiceStore,
		dependencyIndexingStore: dependencyIndexingStore,

		operations: newOperations(observationContext),
		clock:      glock.NewRealClock(),
		logger:     observationContext.Logger,

		metrics:                newResetterMetrics(observationContext),
		janitorMetrics:         newJanitorMetrics(observationContext),
		depencencySyncMetrics:  workerutil.NewMetrics(observationContext, "codeintel_dependency_index_processor"),
		depencencyIndexMetrics: workerutil.NewMetrics(observationContext, "codeintel_dependency_index_queueing"),
	}
}

func (b backgroundJob) WorkerutilStore() dbworkerstore.Store     { return b.workerutilStore }
func (b backgroundJob) DependencySyncStore() dbworkerstore.Store { return b.dependencySyncStore }
func (b backgroundJob) DependencyIndexingStore() dbworkerstore.Store {
	return b.dependencyIndexingStore
}

func (b backgroundJob) InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error) {
	return b.store.InsertDependencyIndexingJob(ctx, uploadID, externalServiceKind, syncTime)
}

// QueueIndexes enqueues a set of index jobs for the following repository and commit. If a non-empty
// configuration is given, it will be used to determine the set of jobs to enqueue. Otherwise, it will
// the configuration will be determined based on the regular index scheduling rules: first read any
// in-repo configuration (e.g., sourcegraph.yaml), then look for any existing in-database configuration,
// finally falling back to the automatically inferred configuration based on the repo contents at the
// target commit.
//
// If the force flag is false, then the presence of an upload or index record for this given repository and commit
// will cause this method to no-op. Note that this is NOT a guarantee that there will never be any duplicate records
// when the flag is false.
func (b backgroundJob) QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []types.Index, err error) {
	ctx, trace, endObservation := b.operations.queueIndex.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.Int("repositoryID", repositoryID),
			otlog.String("rev", rev),
		},
	})
	defer endObservation(1, observation.Args{})

	commitID, err := b.gitserverClient.ResolveRevision(ctx, repositoryID, rev)
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.ResolveRevision")
	}
	commit := string(commitID)
	trace.Log(otlog.String("commit", commit))

	return b.queueIndexForRepositoryAndCommit(ctx, repositoryID, commit, configuration, force, bypassLimit, nil) // trace)
}

// QueueIndexesForPackage enqueues index jobs for a dependency of a recently-processed precise code
// intelligence index.
func (b backgroundJob) QueueIndexesForPackage(ctx context.Context, pkg precise.Package) (err error) {
	ctx, trace, endObservation := b.operations.queueIndexForPackage.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.String("scheme", pkg.Scheme),
			otlog.String("name", pkg.Name),
			otlog.String("version", pkg.Version),
		},
	})
	defer endObservation(1, observation.Args{})

	repoName, revision, ok := InferRepositoryAndRevision(pkg)
	if !ok {
		return nil
	}
	trace.Log(otlog.String("repoName", string(repoName)))
	trace.Log(otlog.String("revision", revision))

	resp, err := b.repoUpdater.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "repoUpdater.EnqueueRepoUpdate")
	}

	commit, err := b.gitserverClient.ResolveRevision(ctx, int(resp.ID), revision)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "gitserverClient.ResolveRevision")
	}

	_, err = b.queueIndexForRepositoryAndCommit(ctx, int(resp.ID), string(commit), "", false, false, nil) // trace)
	return err
}

var (
	overrideScript                           = os.Getenv("SRC_CODEINTEL_INFERENCE_OVERRIDE_SCRIPT")
	maximumIndexJobsPerInferredConfiguration = env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_INDEX_JOBS_PER_INFERRED_CONFIGURATION", 25, "Repositories with a number of inferred auto-index jobs exceeding this threshold will not be auto-indexed.")
)

// InferIndexJobsFromRepositoryStructure collects the result of  InferIndexJobs over all registered recognizers.
func (b backgroundJob) InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]config.IndexJob, error) {
	repoName, err := b.uploadSvc.GetRepoName(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	script, err := b.store.GetInferenceScript(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch inference script from database")
	}
	if script == "" {
		script = overrideScript
	}

	indexes, err := b.inferenceSvc.InferIndexJobs(ctx, api.RepoName(repoName), commit, script)
	if err != nil {
		return nil, err
	}

	if !bypassLimit && len(indexes) > maximumIndexJobsPerInferredConfiguration {
		b.logger.Info("Too many inferred roots. Scheduling no index jobs for repository.", log.Int("repository_id", repositoryID))
		return nil, nil
	}

	return indexes, nil
}

// inferIndexJobsFromRepositoryStructure collects the result of  InferIndexJobHints over all registered recognizers.
func (b backgroundJob) InferIndexJobHintsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error) {
	repoName, err := b.uploadSvc.GetRepoName(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	indexes, err := b.inferenceSvc.InferIndexJobHints(ctx, api.RepoName(repoName), commit, overrideScript)
	if err != nil {
		return nil, err
	}

	return indexes, nil
}

// queueIndexForRepositoryAndCommit determines a set of index jobs to enqueue for the given repository and commit.
//
// If the force flag is false, then the presence of an upload or index record for this given repository and commit
// will cause this method to no-op. Note that this is NOT a guarantee that there will never be any duplicate records
// when the flag is false.
func (b *backgroundJob) queueIndexForRepositoryAndCommit(ctx context.Context, repositoryID int, commit, configuration string, force, bypassLimit bool, trace observation.TraceLogger) ([]types.Index, error) {
	if !force {
		isQueued, err := b.store.IsQueued(ctx, repositoryID, commit)
		if err != nil {
			return nil, errors.Wrap(err, "dbstore.IsQueued")
		}
		if isQueued {
			return nil, nil
		}
	}

	indexes, err := b.getIndexRecords(ctx, repositoryID, commit, configuration, bypassLimit)
	if err != nil {
		return nil, err
	}
	if len(indexes) == 0 {
		return nil, nil
	}

	return b.store.InsertIndexes(ctx, indexes)
}

type configurationFactoryFunc func(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]types.Index, bool, error)

// getIndexRecords determines the set of index records that should be enqueued for the given commit.
// For each repository, we look for index configuration in the following order:
//
//   - supplied explicitly via parameter
//   - in the database
//   - committed to `sourcegraph.yaml` in the repository
//   - inferred from the repository structure
func (b *backgroundJob) getIndexRecords(ctx context.Context, repositoryID int, commit, configuration string, bypassLimit bool) ([]types.Index, error) {
	fns := []configurationFactoryFunc{
		makeExplicitConfigurationFactory(configuration),
		b.getIndexRecordsFromConfigurationInDatabase,
		b.getIndexRecordsFromConfigurationInRepository,
		b.inferIndexRecordsFromRepositoryStructure,
	}

	for _, fn := range fns {
		if indexRecords, ok, err := fn(ctx, repositoryID, commit, bypassLimit); err != nil {
			return nil, err
		} else if ok {
			return indexRecords, nil
		}
	}

	return nil, nil
}

// makeExplicitConfigurationFactory returns a factory that returns a set of index jobs configured
// explicitly via a GraphQL query parameter. If no configuration was supplield then a false valued
// flag is returned.
func makeExplicitConfigurationFactory(configuration string) configurationFactoryFunc {
	logger := log.Scoped("explicitConfigurationFactory", "")
	return func(ctx context.Context, repositoryID int, commit string, _ bool) ([]types.Index, bool, error) {
		if configuration == "" {
			return nil, false, nil
		}

		indexConfiguration, err := config.UnmarshalJSON([]byte(configuration))
		if err != nil {
			// We failed here, but do not try to fall back on another method as having
			// an explicit config supplied via parameter should always take precedence,
			// even if it's broken.
			logger.Warn("Failed to unmarshal index configuration", log.Int("repository_id", repositoryID), log.Error(err))
			return nil, true, nil
		}

		return convertIndexConfiguration(repositoryID, commit, indexConfiguration), true, nil
	}
}

// getIndexRecordsFromConfigurationInDatabase returns a set of index jobs configured via the UI for
// the given repository. If no jobs are configured via the UI then a false valued flag is returned.
func (b *backgroundJob) getIndexRecordsFromConfigurationInDatabase(ctx context.Context, repositoryID int, commit string, _ bool) ([]types.Index, bool, error) {
	indexConfigurationRecord, ok, err := b.store.GetIndexConfigurationByRepositoryID(ctx, repositoryID)
	if err != nil {
		return nil, false, errors.Wrap(err, "dbstore.GetIndexConfigurationByRepositoryID")
	}
	if !ok {
		return nil, false, nil
	}

	indexConfiguration, err := config.UnmarshalJSON(indexConfigurationRecord.Data)
	if err != nil {
		// We failed here, but do not try to fall back on another method as having
		// an explicit config in the database should always take precedence, even
		// if it's broken.
		b.logger.Warn("Failed to unmarshal index configuration", log.Int("repository_id", repositoryID), log.Error(err))
		return nil, true, nil
	}

	return convertIndexConfiguration(repositoryID, commit, indexConfiguration), true, nil
}

// getIndexRecordsFromConfigurationInRepository returns a set of index jobs configured via a committed
// configuration file at the given commit. If no jobs are configured within the repository then a false
// valued flag is returned.
func (b *backgroundJob) getIndexRecordsFromConfigurationInRepository(ctx context.Context, repositoryID int, commit string, _ bool) ([]types.Index, bool, error) {
	isConfigured, err := b.gitserverClient.FileExists(ctx, repositoryID, commit, "sourcegraph.yaml")
	if err != nil {
		return nil, false, errors.Wrap(err, "gitserver.FileExists")
	}
	if !isConfigured {
		return nil, false, nil
	}

	content, err := b.gitserverClient.RawContents(ctx, repositoryID, commit, "sourcegraph.yaml")
	if err != nil {
		return nil, false, errors.Wrap(err, "gitserver.RawContents")
	}

	indexConfiguration, err := config.UnmarshalYAML(content)
	if err != nil {
		// We failed here, but do not try to fall back on another method as having
		// an explicit config in the repository should always take precedence over
		// an auto-inferred configuration, even if it's broken.
		b.logger.Warn("Failed to unmarshal index configuration", log.Int("repository_id", repositoryID), log.Error(err))
		return nil, true, nil
	}

	return convertIndexConfiguration(repositoryID, commit, indexConfiguration), true, nil
}

// inferIndexRecordsFromRepositoryStructure looks at the repository contents at the given commit and
// determines a set of index jobs that are likely to succeed. If no jobs could be inferred then a
// false valued flag is returned.
func (b *backgroundJob) inferIndexRecordsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]types.Index, bool, error) {
	indexJobs, err := b.InferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, bypassLimit)
	if err != nil || len(indexJobs) == 0 {
		return nil, false, err
	}

	return convertInferredConfiguration(repositoryID, commit, indexJobs), true, nil
}

// convertIndexConfiguration converts an index configuration object into a set of index records to be
// inserted into the database.
func convertIndexConfiguration(repositoryID int, commit string, indexConfiguration config.IndexConfiguration) (indexes []types.Index) {
	for _, indexJob := range indexConfiguration.IndexJobs {
		var dockerSteps []types.DockerStep
		for _, dockerStep := range indexConfiguration.SharedSteps {
			dockerSteps = append(dockerSteps, types.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}
		for _, dockerStep := range indexJob.Steps {
			dockerSteps = append(dockerSteps, types.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}

		indexes = append(indexes, types.Index{
			Commit:       commit,
			RepositoryID: repositoryID,
			State:        "queued",
			DockerSteps:  dockerSteps,
			LocalSteps:   indexJob.LocalSteps,
			Root:         indexJob.Root,
			Indexer:      indexJob.Indexer,
			IndexerArgs:  indexJob.IndexerArgs,
			Outfile:      indexJob.Outfile,
		})
	}

	return indexes
}

// convertInferredConfiguration converts a set of index jobs into a set of index records to be inserted
// into the database.
func convertInferredConfiguration(repositoryID int, commit string, indexJobs []config.IndexJob) (indexes []types.Index) {
	for _, indexJob := range indexJobs {
		var dockerSteps []types.DockerStep
		for _, dockerStep := range indexJob.Steps {
			dockerSteps = append(dockerSteps, types.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}

		indexes = append(indexes, types.Index{
			RepositoryID: repositoryID,
			Commit:       commit,
			State:        "queued",
			DockerSteps:  dockerSteps,
			LocalSteps:   indexJob.LocalSteps,
			Root:         indexJob.Root,
			Indexer:      indexJob.Indexer,
			IndexerArgs:  indexJob.IndexerArgs,
			Outfile:      indexJob.Outfile,
		})
	}

	return indexes
}
