package syntactic_indexing

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	matcher := policies.NewMatcher(
		services.GitserverClient,
		policies.IndexingExtractor,
		false,
		true,
	)

	repoSchedulingStore := reposcheduler.NewSyntacticStore(observationCtx, db)
	repoSchedulingSvc := reposcheduler.NewService(repoSchedulingStore)

	return []goroutine.BackgroundRoutine{
		newScheduler(
			observationCtx, repoSchedulingSvc, matcher, *services.PoliciesService, db.Repos(),
		),
	}, nil

}

func newScheduler(
	observationCtx *observation.Context,
	repositorySchedulingService reposcheduler.RepositorySchedulingService,
	policyMatcher autoindexing.PolicyMatcher,
	policiesService policies.Service,
	repoStore database.RepoStore,
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
				repo, _ := repoStore.Get(ctx, repoToIndex.ID)
				fmt.Println(repo.Name)

				offset := 0
				t := true

				policies, totalCount, err := policiesService.GetConfigurationPolicies(ctx, policiesshared.GetConfigurationPoliciesOptions{
					RepositoryID:         int(repoToIndex.ID),
					ForSyntacticIndexing: &t,
					Limit:                config.PolicyBatchSize,
					Offset:               offset,
				})

				if err != nil {
					return err
				}

				fmt.Println("Policies: ", policies, totalCount)
				// if err != nil {
				// 	return errors.Wrap(err, "policySvc.GetConfigurationPolicies")
				// }
				// offset += len(policies)

			}
			// return job.handleScheduler(ctx, config.RepositoryProcessDelay, config.RepositoryBatchSize, config.PolicyBatchSize, config.InferenceConcurrency)
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
