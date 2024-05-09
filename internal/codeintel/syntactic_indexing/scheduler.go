package syntactic_indexing

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/internal"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type syntacticIndexingScheduler struct{}

var _ job.Job = &syntacticIndexingScheduler{}

var config *SchedulerConfig = &SchedulerConfig{}

func NewSyntacticindexingSchedulerJob() job.Job {
	return &syntacticIndexingScheduler{}
}

func (j *syntacticIndexingScheduler) Description() string {
	return ""
}

func (j *syntacticIndexingScheduler) Config() []env.Config {
	return []env.Config{
		config,
	}
}

func (j *syntacticIndexingScheduler) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	rawDB, err := workerdb.InitRawDB(observationCtx)
	if err != nil {
		return nil, err
	}

	db := database.NewDB(observationCtx.Logger, rawDB)
	matcher := policies.NewMatcher(
		services.GitserverClient,
		policies.IndexingExtractor,
		true,
		true,
	)

	repoSchedulingStore := reposcheduler.NewSyntacticStore(observationCtx, db)
	repoSchedulingSvc := reposcheduler.NewService(repoSchedulingStore)

	jobStore, err := jobstore.NewStoreWithDB(observationCtx, rawDB)

	enqueuer := NewIndexEnqueuer(observationCtx, jobStore, repoSchedulingStore, db.Repos(), services.GitserverClient)

	return []goroutine.BackgroundRoutine{
		newScheduler(
			observationCtx,
			repoSchedulingSvc,
			matcher,
			*services.PoliciesService,
			db.Repos(),
			enqueuer,
		),
	}, nil

}

func newScheduler(
	observationCtx *observation.Context,
	repositorySchedulingService reposcheduler.RepositorySchedulingService,
	policyMatcher autoindexing.PolicyMatcher,
	policiesService policies.Service,
	repoStore database.RepoStore,
	enqueuer IndexEnqueuer,
) goroutine.BackgroundRoutine {
	batchOptions := reposcheduler.NewBatchOptions(config.RepositoryProcessDelay, true, &config.PolicyBatchSize, config.RepositoryBatchSize)

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(context.Background()),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			observationCtx.Logger.Info("Launching syntactic indexer...")
			repos, err := repositorySchedulingService.GetRepositoriesForIndexScan(ctx,
				batchOptions, time.Now())

			if err != nil {
				return err
			}

			for _, repoToIndex := range repos {
				repo, _ := repoStore.Get(ctx, api.RepoID(repoToIndex.ID))
				policyIterator := internal.NewPolicyIterator(policiesService, repoToIndex.ID, internal.SyntacticIndexing, config.PolicyBatchSize)

				err := policyIterator.ForEachPoliciesBatch(ctx, func(policies []policiesshared.ConfigurationPolicy) error {

					commitMap, err := policyMatcher.CommitsDescribedByPolicy(ctx, int(repoToIndex.ID), repo.Name, policies, time.Now())
					if err != nil {
						return err
					}

					for commit, policyMatches := range commitMap {
						if len(policyMatches) == 0 {
							continue
						}

						options := EnqueueOptions{force: false, bypassLimit: false}

						// Attempt to queue an index if one does not exist for each of the matching commits
						if _, err := enqueuer.QueueIndexes(ctx, int(repoToIndex.ID), commit, "", options); err != nil {
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
		}),
		goroutine.WithName("codeintel.autoindexing-background-scheduler"),
		goroutine.WithDescription("schedule autoindexing jobs in the background using defined or inferred configurations"),
		goroutine.WithInterval(time.Second*5),
		goroutine.WithOperation(observationCtx.Operation(observation.Op{
			Name:              "codeintel.indexing.HandleIndexSchedule",
			MetricLabelValues: []string{"HandleIndexSchedule"},
			// Metrics:           redMetrics,
			// ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
			// 	if errors.As(err, &inference.LimitError{}) {
			// 		return observation.EmitForNone
			// 	}
			// 	return observation.EmitForDefault
			// },
		})),
	)
}
