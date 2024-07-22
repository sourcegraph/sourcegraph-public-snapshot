package jobselector

import (
	"context"
	"io"
	"os"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type JobSelector struct {
	store           store.Store
	repoStore       database.RepoStore
	inferenceSvc    InferenceService
	gitserverClient gitserver.Client
	logger          log.Logger
}

func NewJobSelector(
	store store.Store,
	repoStore database.RepoStore,
	inferenceSvc InferenceService,
	gitserverClient gitserver.Client,
	logger log.Logger,
) *JobSelector {
	return &JobSelector{
		store:           store,
		repoStore:       repoStore,
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
func (s *JobSelector) InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, localOverrideScript string, bypassLimit bool) (*shared.InferenceResult, error) {
	repo, err := s.repoStore.Get(ctx, api.RepoID(repositoryID))
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
	if localOverrideScript != "" {
		script = localOverrideScript
	}

	if _, canInfer, err := s.store.RepositoryExceptions(ctx, repositoryID); err != nil {
		return nil, err
	} else if !canInfer {
		s.logger.Warn("Auto-indexing job inference for this repo is disabled", log.Int("repositoryID", repositoryID), log.String("repoName", string(repo.Name)))
		return nil, nil
	}

	result, err := s.inferenceSvc.InferIndexJobs(ctx, repo.Name, commit, script)
	if err != nil {
		return nil, err
	}

	if !bypassLimit && len(result.IndexJobs) > MaximumIndexJobsPerInferredConfiguration {
		s.logger.Info("Too many inferred roots. Scheduling no index jobs for repository.", log.Int("repository_id", repositoryID))
		result.IndexJobs = nil
	}

	return result, nil
}

type configurationFactoryFunc func(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]uploadsshared.AutoIndexJob, bool, error)

// GetJobs determines the set of index records that should be enqueued for the given commit.
// For each repository, we look for index configuration in the following order:
//
//   - supplied explicitly via parameter
//   - in the database
//   - committed to `sourcegraph.yaml` in the repository
//   - inferred from the repository structure
func (s *JobSelector) GetJobs(ctx context.Context, repositoryID int, commit, configuration string, bypassLimit bool) ([]uploadsshared.AutoIndexJob, error) {
	if canSchedule, _, err := s.store.RepositoryExceptions(ctx, repositoryID); err != nil {
		return nil, err
	} else if !canSchedule {
		s.logger.Warn("Auto-indexing scheduling for this repo is disabled", log.Int("repositoryID", repositoryID))
		return nil, nil
	}

	fns := []configurationFactoryFunc{
		makeExplicitConfigurationFactory(configuration),
		s.getJobsFromConfigInDatabase,
		s.getJobsFromConfigInRepo,
		s.inferJobsFromRepoContents,
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
	logger := log.Scoped("explicitConfigurationFactory")
	return func(ctx context.Context, repositoryID int, commit string, _ bool) ([]uploadsshared.AutoIndexJob, bool, error) {
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

		return makeQueuedJobs(indexConfiguration.JobSpecs, repositoryID, commit), true, nil
	}
}

// getJobsFromConfigInDatabase returns a set of index jobs configured via the UI for
// the given repository. If no jobs are configured via the UI then a false valued flag is returned.
func (s *JobSelector) getJobsFromConfigInDatabase(ctx context.Context, repositoryID int, commit string, _ bool) ([]uploadsshared.AutoIndexJob, bool, error) {
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

	return makeQueuedJobs(indexConfiguration.JobSpecs, repositoryID, commit), true, nil
}

// getJobsFromConfigInRepo returns a set of index jobs configured via a committed
// configuration file at the given commit. If no jobs are configured within the repository then a false
// valued flag is returned.
func (s *JobSelector) getJobsFromConfigInRepo(ctx context.Context, repositoryID int, commit string, _ bool) ([]uploadsshared.AutoIndexJob, bool, error) {
	repo, err := s.repoStore.Get(ctx, api.RepoID(repositoryID))
	if err != nil {
		return nil, false, err
	}

	r, err := s.gitserverClient.NewFileReader(ctx, repo.Name, api.CommitID(commit), "sourcegraph.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}

		return nil, false, err
	}

	content, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		return nil, false, err
	}

	indexConfiguration, err := config.UnmarshalYAML(content)
	if err != nil {
		// We failed here, but do not try to fall back on another method as having
		// an explicit config in the repository should always take precedence over
		// an auto-inferred configuration, even if it's broken.
		s.logger.Warn("Failed to unmarshal index configuration", log.Int("repository_id", repositoryID), log.Error(err))
		return nil, true, nil
	}

	return makeQueuedJobs(indexConfiguration.JobSpecs, repositoryID, commit), true, nil
}

// inferJobsFromRepoContents looks at the repository contents at the given commit and
// determines a set of index jobs that are likely to succeed. If no jobs could be inferred then a
// false valued flag is returned.
func (s *JobSelector) inferJobsFromRepoContents(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]uploadsshared.AutoIndexJob, bool, error) {
	result, err := s.InferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, "", bypassLimit)
	if err != nil || len(result.IndexJobs) == 0 {
		return nil, false, err
	}

	return makeQueuedJobs(result.IndexJobs, repositoryID, commit), true, nil
}

// TODO: Push api.RepoID and api.CommitID much further up.
func makeQueuedJobs(indexJobs []config.AutoIndexJobSpec, repoID int, commit string) []uploadsshared.AutoIndexJob {
	return genslices.Map(indexJobs, func(job config.AutoIndexJobSpec) uploadsshared.AutoIndexJob {
		return uploadsshared.NewAutoIndexJob(job, api.RepoID(repoID), api.CommitID(commit), uploadsshared.JobStateQueued)
	})
}
