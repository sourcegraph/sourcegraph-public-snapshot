package syntactic_indexing

import (
	"context"
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
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

func NewSyntacticJobScheduler(observationCtx *observation.Context, db *sql.DB) (SyntacticJobScheduler, error) {

	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	database := database.NewDB(observationCtx.Logger, db)
	matcher := policies.NewMatcher(
		services.GitserverClient,
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

	enqueuer := NewIndexEnqueuer(observationCtx, jobStore, repoSchedulingStore, repoStore, services.GitserverClient)

	return &syntacticJobScheduler{
		RepositorySchedulingService: repoSchedulingSvc,
		PolicyMatcher:               matcher,
		PoliciesService:             *services.PoliciesService,
		RepoStore:                   repoStore,
		Enqueuer:                    enqueuer,
		Config:                      config,
	}, nil
}

func (s *syntacticJobScheduler) Schedule(observationCtx *observation.Context, ctx context.Context, currentTime time.Time) error {
	observationCtx.Logger.Info("Launching syntactic indexer...")
	batchOptions := reposcheduler.NewBatchOptions(config.RepositoryProcessDelay, true, &config.PolicyBatchSize, config.RepositoryBatchSize)

	repos, err := s.RepositorySchedulingService.GetRepositoriesForIndexScan(ctx,
		batchOptions, currentTime)

	if err != nil {
		return err
	}

	for _, repoToIndex := range repos {
		repo, _ := s.RepoStore.Get(ctx, api.RepoID(repoToIndex.ID))
		policyIterator := internal.NewPolicyIterator(s.PoliciesService, repoToIndex.ID, internal.SyntacticIndexing, config.PolicyBatchSize)

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
				if _, err := s.Enqueuer.QueueIndexes(ctx, int(repoToIndex.ID), commit, "", options); err != nil {
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
