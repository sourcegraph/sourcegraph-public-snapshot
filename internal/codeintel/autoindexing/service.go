package autoindexing

import (
	"context"
	"os"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ service = (*Service)(nil)

type service interface {
	// Commits
	GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error)
	UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (indexesUpdated int, err error)
	DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (indexesDeleted int, err error)

	// Indexes
	GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) (_ []shared.Index, _ int, err error)
	GetIndexByID(ctx context.Context, id int) (_ shared.Index, _ bool, err error)
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []shared.Index, err error)
	GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error)
	GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error)
	DeleteIndexByID(ctx context.Context, id int) (_ bool, err error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (map[int]int, error)
	QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []shared.Index, err error)
	QueueIndexesForPackage(ctx context.Context, pkg precise.Package) (err error)

	// Index configurations
	GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (_ shared.IndexConfiguration, _ bool, err error)
	InferIndexConfiguration(ctx context.Context, repositoryID int, commit string, bypassLimit bool) (_ *config.IndexConfiguration, hints []config.IndexJobHint, err error)
	UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) (err error)
}

type Service struct {
	store            store.Store
	uploadSvc        shared.UploadService
	gitserverClient  shared.GitserverClient
	repoUpdater      shared.RepoUpdaterClient
	inferenceService shared.InferenceService
	operations       *operations
	logger           log.Logger
}

func newService(
	store store.Store,
	uploadSvc shared.UploadService,
	gitserver shared.GitserverClient,
	repoUpdater shared.RepoUpdaterClient,
	inferenceSvc shared.InferenceService,
	observationContext *observation.Context,
) *Service {
	return &Service{
		store:            store,
		uploadSvc:        uploadSvc,
		gitserverClient:  gitserver,
		repoUpdater:      repoUpdater,
		inferenceService: inferenceSvc,
		operations:       newOperations(observationContext),
		logger:           observationContext.Logger,
	}
}

func (s *Service) GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) (_ []shared.Index, _ int, err error) {
	ctx, _, endObservation := s.operations.getIndexes.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetIndexes(ctx, opts)
}

func (s *Service) GetIndexByID(ctx context.Context, id int) (_ shared.Index, _ bool, err error) {
	ctx, _, endObservation := s.operations.getIndexByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetIndexByID(ctx, id)
}

func (s *Service) GetIndexesByIDs(ctx context.Context, ids ...int) (_ []shared.Index, err error) {
	ctx, _, endObservation := s.operations.getIndexesByIDs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetIndexesByIDs(ctx, ids...)
}

func (s *Service) GetRecentIndexesSummary(ctx context.Context, repositoryID int) (summaries []shared.IndexesWithRepositoryNamespace, err error) {
	ctx, _, endObservation := s.operations.getRecentIndexesSummary.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetRecentIndexesSummary(ctx, repositoryID)
}

func (s *Service) GetLastIndexScanForRepository(ctx context.Context, repositoryID int) (_ *time.Time, err error) {
	ctx, _, endObservation := s.operations.getLastIndexScanForRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetLastIndexScanForRepository(ctx, repositoryID)
}

func (s *Service) DeleteIndexByID(ctx context.Context, id int) (_ bool, err error) {
	ctx, _, endObservation := s.operations.deleteIndexByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteIndexByID(ctx, id)
}

func (s *Service) DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (_ map[int]int, err error) {
	ctx, _, endObservation := s.operations.deleteIndexesWithoutRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteIndexesWithoutRepository(ctx, now)
}

func (s *Service) GetStaleSourcedCommits(ctx context.Context, minimumTimeSinceLastCheck time.Duration, limit int, now time.Time) (_ []shared.SourcedCommits, err error) {
	ctx, _, endObservation := s.operations.getStaleSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetStaleSourcedCommits(ctx, minimumTimeSinceLastCheck, limit, now)
}

func (s *Service) UpdateSourcedCommits(ctx context.Context, repositoryID int, commit string, now time.Time) (indexesUpdated int, err error) {
	ctx, _, endObservation := s.operations.updateSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateSourcedCommits(ctx, repositoryID, commit, now)
}

func (s *Service) DeleteSourcedCommits(ctx context.Context, repositoryID int, commit string, maximumCommitLag time.Duration, now time.Time) (indexesDeleted int, err error) {
	ctx, _, endObservation := s.operations.deleteSourcedCommits.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteSourcedCommits(ctx, repositoryID, commit, maximumCommitLag)
}

func (s *Service) GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (_ shared.IndexConfiguration, _ bool, err error) {
	ctx, _, endObservation := s.operations.getIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetIndexConfigurationByRepositoryID(ctx, repositoryID)
}

// InferIndexConfiguration looks at the repository contents at the latest commit on the default branch of the given
// repository and determines an index configuration that is likely to succeed.
func (s *Service) InferIndexConfiguration(ctx context.Context, repositoryID int, commit string, bypassLimit bool) (_ *config.IndexConfiguration, hints []config.IndexJobHint, err error) {
	ctx, trace, endObservation := s.operations.inferIndexConfiguration.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.Int("repositoryID", repositoryID),
		},
	})
	defer endObservation(1, observation.Args{})

	if commit == "" {
		var ok bool
		commit, ok, err = s.gitserverClient.Head(ctx, repositoryID)
		if err != nil || !ok {
			return nil, nil, errors.Wrapf(err, "gitserver.Head: error resolving HEAD for %d", repositoryID)
		}
	} else {
		exists, err := s.gitserverClient.CommitExists(ctx, repositoryID, commit)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "gitserver.CommitExists: error checking %s for %d", commit, repositoryID)
		}

		if !exists {
			return nil, nil, errors.Newf("revision %s not found for %d", commit, repositoryID)
		}
	}
	trace.Log(otlog.String("commit", commit))

	indexJobs, err := s.inferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, bypassLimit)
	if err != nil {
		return nil, nil, err
	}

	indexJobHints, err := s.inferIndexJobHintsFromRepositoryStructure(ctx, repositoryID, commit)
	if err != nil {
		return nil, nil, err
	}

	if len(indexJobs) == 0 {
		return nil, indexJobHints, nil
	}

	return &config.IndexConfiguration{
		IndexJobs: indexJobs,
	}, indexJobHints, nil
}

func (s *Service) UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) (err error) {
	ctx, _, endObservation := s.operations.updateIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.UpdateIndexConfigurationByRepositoryID(ctx, repositoryID, data)
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
func (s *Service) QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) (_ []shared.Index, err error) {
	ctx, trace, endObservation := s.operations.queueIndex.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.Int("repositoryID", repositoryID),
		},
	})
	defer endObservation(1, observation.Args{})

	commitID, err := s.gitserverClient.ResolveRevision(ctx, repositoryID, rev)
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.ResolveRevision")
	}
	commit := string(commitID)
	trace.Log(otlog.String("commit", commit))

	return s.queueIndexForRepositoryAndCommit(ctx, repositoryID, commit, configuration, force, bypassLimit, nil) // trace)
}

// QueueIndexesForPackage enqueues index jobs for a dependency of a recently-processed precise code
// intelligence index.
func (s *Service) QueueIndexesForPackage(ctx context.Context, pkg precise.Package) (err error) {
	ctx, trace, endObservation := s.operations.queueIndexForPackage.With(ctx, &err, observation.Args{
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

	resp, err := s.repoUpdater.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "repoUpdater.EnqueueRepoUpdate")
	}

	commit, err := s.gitserverClient.ResolveRevision(ctx, int(resp.ID), revision)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}

		return errors.Wrap(err, "gitserverClient.ResolveRevision")
	}

	_, err = s.queueIndexForRepositoryAndCommit(ctx, int(resp.ID), string(commit), "", false, false, nil) // trace)
	return err
}

// queueIndexForRepositoryAndCommit determines a set of index jobs to enqueue for the given repository and commit.
//
// If the force flag is false, then the presence of an upload or index record for this given repository and commit
// will cause this method to no-op. Note that this is NOT a guarantee that there will never be any duplicate records
// when the flag is false.
func (s *Service) queueIndexForRepositoryAndCommit(ctx context.Context, repositoryID int, commit, configuration string, force, bypassLimit bool, trace observation.TraceLogger) ([]shared.Index, error) {
	if !force {
		isQueued, err := s.store.IsQueued(ctx, repositoryID, commit)
		if err != nil {
			return nil, errors.Wrap(err, "dbstore.IsQueued")
		}
		if isQueued {
			return nil, nil
		}
	}

	indexes, err := s.getIndexRecords(ctx, repositoryID, commit, configuration, bypassLimit)
	if err != nil {
		return nil, err
	}
	if len(indexes) == 0 {
		return nil, nil
	}

	return s.store.InsertIndexes(ctx, indexes)
}

var overrideScript = os.Getenv("SRC_CODEINTEL_INFERENCE_OVERRIDE_SCRIPT")

// inferIndexJobsFromRepositoryStructure collects the result of  InferIndexJobs over all registered recognizers.
func (s *Service) inferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]config.IndexJob, error) {
	repoName, err := s.uploadSvc.GetRepoName(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	indexes, err := s.inferenceService.InferIndexJobs(ctx, api.RepoName(repoName), commit, overrideScript)
	if err != nil {
		return nil, err
	}

	if !bypassLimit && len(indexes) > maximumIndexJobsPerInferredConfiguration {
		s.logger.Info("Too many inferred roots. Scheduling no index jobs for repository.", log.Int("repository_id", repositoryID))
		return nil, nil
	}

	return indexes, nil
}

// inferIndexJobsFromRepositoryStructure collects the result of  InferIndexJobHints over all registered recognizers.
func (s *Service) inferIndexJobHintsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error) {
	repoName, err := s.uploadSvc.GetRepoName(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	indexes, err := s.inferenceService.InferIndexJobHints(ctx, api.RepoName(repoName), commit, overrideScript)
	if err != nil {
		return nil, err
	}

	return indexes, nil
}

type configurationFactoryFunc func(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]shared.Index, bool, error)

// getIndexRecords determines the set of index records that should be enqueued for the given commit.
// For each repository, we look for index configuration in the following order:
//
//   - supplied explicitly via parameter
//   - in the database
//   - committed to `sourcegraph.yaml` in the repository
//   - inferred from the repository structure
func (s *Service) getIndexRecords(ctx context.Context, repositoryID int, commit, configuration string, bypassLimit bool) ([]shared.Index, error) {
	fns := []configurationFactoryFunc{
		makeExplicitConfigurationFactory(configuration),
		s.getIndexRecordsFromConfigurationInDatabase,
		s.getIndexRecordsFromConfigurationInRepository,
		s.inferIndexRecordsFromRepositoryStructure,
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
	return func(ctx context.Context, repositoryID int, commit string, _ bool) ([]shared.Index, bool, error) {
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
func (s *Service) getIndexRecordsFromConfigurationInDatabase(ctx context.Context, repositoryID int, commit string, _ bool) ([]shared.Index, bool, error) {
	indexConfigurationRecord, ok, err := s.store.GetIndexConfigurationByRepositoryID(ctx, repositoryID)
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
		s.logger.Warn("Failed to unmarshal index configuration", log.Int("repository_id", repositoryID), log.Error(err))
		return nil, true, nil
	}

	return convertIndexConfiguration(repositoryID, commit, indexConfiguration), true, nil
}

// getIndexRecordsFromConfigurationInRepository returns a set of index jobs configured via a committed
// configuration file at the given commit. If no jobs are configured within the repository then a false
// valued flag is returned.
func (s *Service) getIndexRecordsFromConfigurationInRepository(ctx context.Context, repositoryID int, commit string, _ bool) ([]shared.Index, bool, error) {
	isConfigured, err := s.gitserverClient.FileExists(ctx, repositoryID, commit, "sourcegraph.yaml")
	if err != nil {
		return nil, false, errors.Wrap(err, "gitserver.FileExists")
	}
	if !isConfigured {
		return nil, false, nil
	}

	content, err := s.gitserverClient.RawContents(ctx, repositoryID, commit, "sourcegraph.yaml")
	if err != nil {
		return nil, false, errors.Wrap(err, "gitserver.RawContents")
	}

	indexConfiguration, err := config.UnmarshalYAML(content)
	if err != nil {
		// We failed here, but do not try to fall back on another method as having
		// an explicit config in the repository should always take precedence over
		// an auto-inferred configuration, even if it's broken.
		s.logger.Warn("Failed to unmarshal index configuration", log.Int("repository_id", repositoryID), log.Error(err))
		return nil, true, nil
	}

	return convertIndexConfiguration(repositoryID, commit, indexConfiguration), true, nil
}

// inferIndexRecordsFromRepositoryStructure looks at the repository contents at the given commit and
// determines a set of index jobs that are likely to succeed. If no jobs could be inferred then a
// false valued flag is returned.
func (s *Service) inferIndexRecordsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]shared.Index, bool, error) {
	indexJobs, err := s.inferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, bypassLimit)
	if err != nil || len(indexJobs) == 0 {
		return nil, false, err
	}

	return convertInferredConfiguration(repositoryID, commit, indexJobs), true, nil
}

// convertIndexConfiguration converts an index configuration object into a set of index records to be
// inserted into the database.
func convertIndexConfiguration(repositoryID int, commit string, indexConfiguration config.IndexConfiguration) (indexes []shared.Index) {
	for _, indexJob := range indexConfiguration.IndexJobs {
		var dockerSteps []shared.DockerStep
		for _, dockerStep := range indexConfiguration.SharedSteps {
			dockerSteps = append(dockerSteps, shared.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}
		for _, dockerStep := range indexJob.Steps {
			dockerSteps = append(dockerSteps, shared.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}

		indexes = append(indexes, shared.Index{
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
func convertInferredConfiguration(repositoryID int, commit string, indexJobs []config.IndexJob) (indexes []shared.Index) {
	for _, indexJob := range indexJobs {
		var dockerSteps []shared.DockerStep
		for _, dockerStep := range indexJob.Steps {
			dockerSteps = append(dockerSteps, shared.DockerStep{
				Root:     dockerStep.Root,
				Image:    dockerStep.Image,
				Commands: dockerStep.Commands,
			})
		}

		indexes = append(indexes, shared.Index{
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
