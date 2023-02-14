package autoindexing

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/regexp"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/enqueuer"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/jobselector"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Service struct {
	store           store.Store
	inferenceSvc    InferenceService
	repoUpdater     RepoUpdaterClient
	gitserverClient GitserverClient
	symbolsClient   *symbols.Client
	indexEnqueuer   *enqueuer.IndexEnqueuer
	jobSelector     *jobselector.JobSelector
	logger          log.Logger
	operations      *operations
}

func newService(
	observationCtx *observation.Context,
	store store.Store,
	inferenceSvc InferenceService,
	repoUpdater RepoUpdaterClient,
	gitserver GitserverClient,
	symbolsClient *symbols.Client,
) *Service {
	// NOTE - this should go up a level in init.go.
	// Not going to do this now so that we don't blow up all of the
	// tests (which have pretty good coverage of the whole service).
	// We should rewrite/transplant tests to the closest package that
	// provides that behavior and then mock the dependencies in the
	// glue packages.

	jobSelector := jobselector.NewJobSelector(
		store,
		inferenceSvc,
		gitserver,
		log.Scoped("autoindexing job selector", ""),
	)

	indexEnqueuer := enqueuer.NewIndexEnqueuer(
		observationCtx,
		store,
		repoUpdater,
		gitserver,
		jobSelector,
	)

	return &Service{
		store:           store,
		inferenceSvc:    inferenceSvc,
		repoUpdater:     repoUpdater,
		gitserverClient: gitserver,
		symbolsClient:   symbolsClient,
		indexEnqueuer:   indexEnqueuer,
		jobSelector:     jobSelector,
		logger:          observationCtx.Logger,
		operations:      newOperations(observationCtx),
	}
}

func (s *Service) GetIndexes(ctx context.Context, opts shared.GetIndexesOptions) (_ []types.Index, _ int, err error) {
	ctx, _, endObservation := s.operations.getIndexes.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetIndexes(ctx, opts)
}

func (s *Service) GetIndexByID(ctx context.Context, id int) (_ types.Index, _ bool, err error) {
	ctx, _, endObservation := s.operations.getIndexByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetIndexByID(ctx, id)
}

func (s *Service) GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error) {
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

func (s *Service) DeleteIndexes(ctx context.Context, opts shared.DeleteIndexesOptions) (err error) {
	ctx, _, endObservation := s.operations.deleteIndexes.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.DeleteIndexes(ctx, opts)
}

func (s *Service) ReindexIndexByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.reindexIndexByID.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.ReindexIndexByID(ctx, id)
}

func (s *Service) ReindexIndexes(ctx context.Context, opts shared.ReindexIndexesOptions) (err error) {
	ctx, _, endObservation := s.operations.reindexIndexes.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.ReindexIndexes(ctx, opts)
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
	trace.AddEvent("found", attribute.String("commit", commit))

	indexJobs, err := s.InferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, bypassLimit)
	if err != nil {
		return nil, nil, err
	}

	indexJobHints, err := s.InferIndexJobHintsFromRepositoryStructure(ctx, repositoryID, commit)
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

func (s *Service) GetUnsafeDB() database.DB {
	return s.store.GetUnsafeDB()
}

func (s *Service) ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error) {
	return s.gitserverClient.ListFiles(ctx, repositoryID, commit, pattern)
}

func (s *Service) GetSupportedByCtags(ctx context.Context, filepath string, repoName api.RepoName) (bool, string, error) {
	mappings, err := s.symbolsClient.ListLanguageMappings(ctx, repoName)
	if err != nil {
		return false, "", err
	}

	for language, globs := range mappings {
		for _, glob := range globs {
			if glob.Match(filepath) {
				return true, language, nil
			}
		}
	}

	return false, "", nil
}

func (s *Service) SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error) {
	ctx, _, endObservation := s.operations.setRequestLanguageSupport.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{otlog.Int("userID", userID), otlog.String("language", language)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.SetRequestLanguageSupport(ctx, userID, language)
}

func (s *Service) GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error) {
	ctx, _, endObservation := s.operations.getLanguagesRequestedBy.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{otlog.Int("userID", userID)},
	})
	defer endObservation(1, observation.Args{})

	return s.store.GetLanguagesRequestedBy(ctx, userID)
}

func (s *Service) GetListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error) {
	ctx, _, endObservation := s.operations.getListTags.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{otlog.String("repo", string(repo)), otlog.String("commitObjs", fmt.Sprintf("%v", commitObjs))},
	})
	defer endObservation(1, observation.Args{})

	return s.gitserverClient.ListTags(ctx, repo, commitObjs...)
}

func (s *Service) QueueRepoRev(ctx context.Context, repositoryID int, rev string) (err error) {
	ctx, _, endObservation := s.operations.queueRepoRev.With(ctx, &err, observation.Args{
		LogFields: []otlog.Field{
			otlog.Int("repositoryID", repositoryID),
			otlog.String("rev", rev),
		},
	})
	defer endObservation(1, observation.Args{})

	return s.store.QueueRepoRev(ctx, repositoryID, rev)
}

func (s *Service) SetInferenceScript(ctx context.Context, script string) (err error) {
	ctx, _, endObservation := s.operations.setInferenceScript.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.SetInferenceScript(ctx, script)
}

func (s *Service) GetInferenceScript(ctx context.Context) (script string, err error) {
	ctx, _, endObservation := s.operations.getInferenceScript.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.store.GetInferenceScript(ctx)
}

func (s *Service) QueueIndexes(ctx context.Context, repositoryID int, rev, configuration string, force, bypassLimit bool) ([]types.Index, error) {
	return s.indexEnqueuer.QueueIndexes(ctx, repositoryID, rev, configuration, force, bypassLimit)
}

func (s *Service) QueueIndexesForPackage(ctx context.Context, pkg dependencies.MinimialVersionedPackageRepo, assumeSynced bool) (err error) {
	return s.indexEnqueuer.QueueIndexesForPackage(ctx, pkg, assumeSynced)
}

func (s *Service) InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, bypassLimit bool) ([]config.IndexJob, error) {
	return s.jobSelector.InferIndexJobsFromRepositoryStructure(ctx, repositoryID, commit, bypassLimit)
}

func (s *Service) InferIndexJobHintsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string) ([]config.IndexJobHint, error) {
	return s.jobSelector.InferIndexJobHintsFromRepositoryStructure(ctx, repositoryID, commit)
}

func (s *Service) Summary(ctx context.Context) (shared.Summary, error) {
	return s.store.Summary(ctx)
}
