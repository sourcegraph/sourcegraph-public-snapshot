package jobselector

import (
	"context"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type JobSelector struct {
	store           store.Store
	uploadSvc       UploadService
	inferenceSvc    InferenceService
	gitserverClient GitserverClient
	logger          log.Logger
}

func NewJobSelector(
	store store.Store,
	uploadSvc UploadService,
	inferenceSvc InferenceService,
	gitserverClient GitserverClient,
	logger log.Logger,
) *JobSelector {
	return &JobSelector{
		store:           store,
		uploadSvc:       uploadSvc,
		inferenceSvc:    inferenceSvc,
		gitserverClient: gitserverClient,
		logger:          logger,
	}
}

var (
	overrideScript                           = os.Getenv("SRC_CODEINTEL_INFERENCE_OVERRIDE_SCRIPT")
	MaximumIndexJobsPerInferredConfiguration = env.MustGetInt("PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_INDEX_JOBS_PER_INFERRED_CONFIGURATION", 50, "Repositories with a number of inferred auto-index jobs exceeding this threshold will not be auto-indexed.")
)

// InferIndexJobsFromRepositoryStructure collects the result of InferIndexJobs over all registered recognizers.
func (s *JobSelector) InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]config.IndexJob, error) {
	repoName, err := s.uploadSvc.GetRepoName(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	script, err := s.store.GetInferenceScript(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch inference script from database")
	}
	if script == "" {
		script = overrideScript
	}

	indexes, err := s.inferenceSvc.InferIndexJobs(ctx, api.RepoName(repoName), commit, script)
	if err != nil {
		return nil, err
	}

	if !bypassLimit && len(indexes) > MaximumIndexJobsPerInferredConfiguration {
		s.logger.Info("Too many inferred roots. Scheduling no index jobs for repository.", log.Int("repository_id", repositoryID))
		return nil, nil
	}

	return indexes, nil
}

// inferIndexJobsFromRepositoryStructure collects the result of  InferIndexJobHints over all registered recognizers.
func (s *JobSelector) InferIndexJobHintsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error) {
	repoName, err := s.uploadSvc.GetRepoName(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	indexes, err := s.inferenceSvc.InferIndexJobHints(ctx, api.RepoName(repoName), commit, overrideScript)
	if err != nil {
		return nil, err
	}

	return indexes, nil
}

type configurationFactoryFunc func(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]types.Index, bool, error)

// GetIndexRecords determines the set of index records that should be enqueued for the given commit.
// For each repository, we look for index configuration in the following order:
//
//   - supplied explicitly via parameter
//   - in the database
//   - committed to `sourcegraph.yaml` in the repository
//   - inferred from the repository structure
func (s *JobSelector) GetIndexRecords(ctx context.Context, repositoryID int, commit, configuration string, bypassLimit bool) ([]types.Index, error) {
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
func (s *JobSelector) getIndexRecordsFromConfigurationInDatabase(ctx context.Context, repositoryID int, commit string, _ bool) ([]types.Index, bool, error) {
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
func (s *JobSelector) getIndexRecordsFromConfigurationInRepository(ctx context.Context, repositoryID int, commit string, _ bool) ([]types.Index, bool, error) {
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
func (s *JobSelector) inferIndexRecordsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]types.Index, bool, error) {
	indexJobs, err := s.InferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, bypassLimit)
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
			Commit:           commit,
			RepositoryID:     repositoryID,
			State:            "queued",
			DockerSteps:      dockerSteps,
			LocalSteps:       indexJob.LocalSteps,
			Root:             indexJob.Root,
			Indexer:          indexJob.Indexer,
			IndexerArgs:      indexJob.IndexerArgs,
			Outfile:          indexJob.Outfile,
			RequestedEnvVars: indexJob.RequestedEnvVars,
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
			RepositoryID:     repositoryID,
			Commit:           commit,
			State:            "queued",
			DockerSteps:      dockerSteps,
			LocalSteps:       indexJob.LocalSteps,
			Root:             indexJob.Root,
			Indexer:          indexJob.Indexer,
			IndexerArgs:      indexJob.IndexerArgs,
			Outfile:          indexJob.Outfile,
			RequestedEnvVars: indexJob.RequestedEnvVars,
		})
	}

	return indexes
}
