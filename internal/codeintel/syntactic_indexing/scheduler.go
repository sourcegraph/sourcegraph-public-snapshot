package syntactic_indexing

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/internal"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SyntacticJobScheduler interface {
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

func NewSyntacticJobScheduler(repoSchedulingSvc reposcheduler.RepositorySchedulingService,
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

func BootstrapSyntacticJobScheduler(observationCtx *observation.Context, frontendDB database.DB, codeIntelDB codeintelshared.CodeIntelDB) (SyntacticJobScheduler, error) {
	gitserverClient := gitserver.NewClient("codeintel-syntactic-indexing")

	uploadsSvc := uploads.NewService(observationCtx, frontendDB, codeIntelDB, gitserverClient.Scoped("uploads"))
	policiesSvc := policies.NewService(observationCtx, frontendDB, uploadsSvc, gitserverClient.Scoped("policies"))

	matcher := policies.NewMatcher(
		gitserverClient,
		policies.IndexingExtractor,
		true,
		true,
	)

	repoSchedulingStore := reposcheduler.NewSyntacticStore(observationCtx, frontendDB)
	repoSchedulingSvc := reposcheduler.NewService(repoSchedulingStore)

	jobStore, err := jobstore.NewStoreWithDB(observationCtx, frontendDB)
	if err != nil {
		return nil, err
	}

	repoStore := frontendDB.Repos()

	enqueuer := NewIndexEnqueuer(observationCtx, jobStore, repoSchedulingStore, repoStore)

	return NewSyntacticJobScheduler(repoSchedulingSvc, *matcher, *policiesSvc, repoStore, enqueuer, *schedulerConfig)
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

	commitsToSchedule := make(map[api.RepoID]collections.Set[api.CommitID])
	enqueueOptions := EnqueueOptions{force: false}

	var allErrors error

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
				if commits := commitsToSchedule[repo.ID]; commits != nil {
					commits.Add(api.CommitID(commit))
				} else {
					commitsToSchedule[repo.ID] = collections.NewSet(api.CommitID(commit))
				}
			}

			return nil
		})

		if err != nil {
			allErrors = errors.Append(allErrors, errors.Newf("Failed to discover commits eligible for syntactic indexing for repo [%s]: %v", repo.Name, err))
		}
	}

	for repoId, commits := range commitsToSchedule {
		for _, commitId := range commits.Values() {
			if _, err := s.Enqueuer.QueueIndexingJobs(ctx, repoId, commitId, enqueueOptions); err != nil {
				allErrors = errors.Append(allErrors, errors.Newf("Failed to schedule syntactic indexing of repo [ID=%s], commit [%s]: %v", repoId, commitId, err))
			}
		}
	}

	return allErrors
}
