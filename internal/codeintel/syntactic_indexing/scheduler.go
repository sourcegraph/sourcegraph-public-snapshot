package syntactic_indexing

import (
	"context"
	"database/sql"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/internal"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SyntacticJobScheduler interface {
	// TODO: make it return job that were queued successfully?
	Schedule(observationCtx *observation.Context, ctx context.Context, currentTime time.Time) error
	GetConfig() *SchedulerConfig
}

type syntacticJobScheduler struct {
	RepositorySchedulingService reposcheduler.RepositorySchedulingService
	PolicyMatcher               autoindexing.PolicyMatcher
	PoliciesService             policies.Service
	RepoStore                   database.RepoStore
	Enqueuer                    IndexEnqueuer
	Config                      *SchedulerConfig
}

var _ SyntacticJobScheduler = &syntacticJobScheduler{}

func NewSyntaticJobScheduler(repoSchedulingSvc reposcheduler.RepositorySchedulingService,
	policyMatcher policies.Matcher, policiesSvc policies.Service,
	repoStore database.RepoStore, enqueuer IndexEnqueuer, config SchedulerConfig) (SyntacticJobScheduler, error) {

	return &syntacticJobScheduler{
		RepositorySchedulingService: repoSchedulingSvc,
		PolicyMatcher:               &policyMatcher,
		PoliciesService:             policiesSvc,
		RepoStore:                   repoStore,
		Enqueuer:                    enqueuer,
		Config:                      &config,
	}, nil
}

func BootstrapSyntacticJobScheduler(observationCtx *observation.Context, db *sql.DB) (SyntacticJobScheduler, error) {
	database := database.NewDB(observationCtx.Logger, db)

	gitserverClient := gitserver.NewClient("codeintel-syntactic-indexing")

	codeIntelDB := codeintelshared.NewCodeIntelDB(observationCtx.Logger, db)

	uploadsSvc := uploads.NewService(observationCtx, database, codeIntelDB, gitserverClient.Scoped("uploads"))
	policiesSvc := policies.NewService(observationCtx, database, uploadsSvc, gitserverClient.Scoped("policies"))

	schedulerConfig.Load()

	matcher := policies.NewMatcher(
		gitserverClient,
		policies.IndexingExtractor,
		true,
		true,
	)

	repoSchedulingStore := reposcheduler.NewSyntacticStore(observationCtx, database)
	repoSchedulingSvc := reposcheduler.NewService(repoSchedulingStore)

	jobStore, err := jobstore.NewStoreWithDB(observationCtx, db)
	if err != nil {
		return nil, err
	}

	repoStore := database.Repos()

	enqueuer := NewIndexEnqueuer(observationCtx, jobStore, repoSchedulingStore, repoStore, gitserverClient)

	return NewSyntaticJobScheduler(repoSchedulingSvc, *matcher, *policiesSvc, repoStore, enqueuer, *schedulerConfig)
}

// GetConfig implements SyntacticJobScheduler.
func (s *syntacticJobScheduler) GetConfig() *SchedulerConfig {
	return s.Config
}

func (s *syntacticJobScheduler) Schedule(observationCtx *observation.Context, ctx context.Context, currentTime time.Time) error {
	batchOptions := reposcheduler.NewBatchOptions(schedulerConfig.RepositoryProcessDelay, true, &schedulerConfig.PolicyBatchSize, schedulerConfig.RepositoryBatchSize)

	repos, err := s.RepositorySchedulingService.GetRepositoriesForIndexScan(ctx,
		batchOptions, currentTime)

	if err != nil {
		return err
	}

	for _, repoToIndex := range repos {
		repo, _ := s.RepoStore.Get(ctx, api.RepoID(repoToIndex.ID))
		policyIterator := internal.NewPolicyIterator(s.PoliciesService, repoToIndex.ID, internal.SyntacticIndexing, schedulerConfig.PolicyBatchSize)

		err := policyIterator.ForEachPoliciesBatch(ctx, func(policies []policiesshared.ConfigurationPolicy) error {

			commitMap, err := s.PolicyMatcher.CommitsDescribedByPolicy(ctx, int(repoToIndex.ID), repo.Name, policies, currentTime)

			if err != nil {
				return err
			}

			for commit, policyMatches := range commitMap {
				if len(policyMatches) == 0 {
					continue
				}

				options := EnqueueOptions{force: false, bypassLimit: false}

				// Attempt to queue an index if one does not exist for each of the matching commits
				if _, err := s.Enqueuer.QueueIndexes(ctx, int(repoToIndex.ID), commit, options); err != nil {
					if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
						continue
					}

					return errors.Wrap(err, "indexEnqueuer.QueueIndexes")
				}
			}

			return nil

		})

		if err != nil {
			return err
		}

	}
	return nil
}
