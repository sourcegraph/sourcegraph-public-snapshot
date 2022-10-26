package background

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/semaphore"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	policiesshared "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type indexSchedulerOperations struct {
	handleIndexScheduler *observation.Operation
	handleRepoScheduler  *observation.Operation
}

func newOperations(observationContext *observation.Context) *indexSchedulerOperations {
	observationContext = &observation.Context{
		Logger:     observationContext.Logger,
		Tracer:     observationContext.Tracer,
		Registerer: observationContext.Registerer,
		HoneyDataset: &honey.Dataset{
			Name: "codeintel-index-scheduler",
		},
	}

	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing_background",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	return &indexSchedulerOperations{
		handleIndexScheduler: observationContext.Operation(observation.Op{
			Name:              "codeintel.indexing.HandleIndexSchedule",
			MetricLabelValues: []string{"HandleIndexSchedule"},
			Metrics:           m,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				if errors.As(err, &inference.LimitError{}) {
					return observation.EmitForDefault.Without(observation.EmitForMetrics)
				}
				return observation.EmitForDefault
			},
		}),
		handleRepoScheduler: observationContext.Operation(observation.Op{
			Name: "codeintel.indexing.HandleRepoSchedule",
		}),
	}
}

func (b *backgroundJob) NewRepoIndexingScheduler(
	interval time.Duration,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	policyBatchSize int,
	inferenceConcurrency int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.handleScheduler(policy.WithShouldTrace(ctx, true), repositoryProcessDelay, repositoryBatchSize, policyBatchSize, inferenceConcurrency)
	}))
}

func (b backgroundJob) handleScheduler(
	ctx context.Context,
	repositoryProcessDelay time.Duration,
	repositoryBatchSize int,
	policyBatchSize int,
	inferenceConcurrency int,
) (err error) {
	if !autoIndexingEnabled() {
		return nil
	}

	fmt.Println("STARTING HANDLE")

	ctx, traceLog, endObservation := b.indexSchedulerOperations.handleIndexScheduler.With(ctx, &err, observation.Args{})
	defer func() {
		fmt.Println("FINISHED HANDLE")
		endObservation(1, observation.Args{})
	}()

	var repositoryMatchLimit *int
	if val := conf.CodeIntelAutoIndexingPolicyRepositoryMatchLimit(); val != -1 {
		repositoryMatchLimit = &val
	}

	// Get the batch of repositories that we'll handle in this invocation of the periodic goroutine. This
	// set should contain repositories that have yet to be updated, or that have been updated least recently.
	// This allows us to update every repository reliably, even if it takes a long time to process through
	// the backlog.
	repositories, err := b.uploadSvc.GetRepositoriesForIndexScan(
		ctx,
		"lsif_last_index_scan",
		"last_index_scan_at",
		repositoryProcessDelay,
		conf.CodeIntelAutoIndexingAllowGlobalPolicies(),
		repositoryMatchLimit,
		repositoryBatchSize,
		time.Now(),
	)

	traceLog.SetAttributes(
		attribute.Int("repoCandidates", len(repositories)),
		attribute.Int("repoMatchLimit", intp(repositoryMatchLimit)),
	)

	if err != nil {
		return errors.Wrap(err, "uploadSvc.GetRepositoriesForIndexScan")
	}
	if len(repositories) == 0 {
		// All repositories updated recently enough
		return nil
	}

	now := timeutil.Now()

	// In parallel enqueue all the repos.
	var (
		sema  = semaphore.NewWeighted(int64(inferenceConcurrency))
		errs  error
		errMu sync.Mutex
	)

	for _, repositoryID := range repositories {
		fmt.Println("STARTING REPO HANDLE")
		var repositoryErr error
		ctx, traceLog, endObservation := b.indexSchedulerOperations.handleRepoScheduler.With(ctx, &repositoryErr, observation.Args{
			LogFields: []log.Field{log.Int("repoID", repositoryID)},
		})

		semWait := time.Now()
		if err := sema.Acquire(ctx, 1); err != nil {
			endObservation(1, observation.Args{})
			return err
		}
		traceLog.SetAttributes(attribute.Int64("semaphoreWaitMs", time.Since(semWait).Milliseconds()))

		go func(repositoryID int) {
			defer func() {
				fmt.Println("FINISHED REPO HANDLE")
				sema.Release(1)
				endObservation(1, observation.Args{})
			}()
			if repositoryErr = b.handleRepository(ctx, repositoryID, policyBatchSize, now, traceLog); repositoryErr != nil {
				errMu.Lock()
				errs = errors.Append(errs, repositoryErr)
				errMu.Unlock()
			}
		}(repositoryID)
	}

	if err := sema.Acquire(ctx, int64(inferenceConcurrency)); err != nil {
		return errors.Wrap(err, "acquiring semaphore")
	}

	return errs
}

func (b backgroundJob) handleRepository(ctx context.Context, repositoryID, policyBatchSize int, now time.Time, traceLog observation.TraceLogger) error {
	var (
		// tracks the amount of commits that had >0 policy matches from this repository,
		// and how many of those had a job successfully enqueued for
		uniqueCommits = make(map[string]struct{})
		jobsQueued    int
		// store pagination
		offset = 0
	)

	defer func() {
		traceLog.SetAttributes(
			attribute.Int("indexJobsEnqueued", jobsQueued),
			attribute.Int("commitsMatched", len(uniqueCommits)),
			attribute.Int("policiesMatched", offset),
		)
	}()

	for {
		// Retrieve the set of configuration policies that affect indexing for this repository.
		policies, totalCount, err := b.policiesSvc.GetConfigurationPolicies(ctx, policiesshared.GetConfigurationPoliciesOptions{
			RepositoryID: repositoryID,
			ForIndexing:  true,
			Limit:        policyBatchSize,
			Offset:       offset,
		})
		if err != nil {
			return errors.Wrap(err, "policySvc.GetConfigurationPolicies")
		}
		offset += len(policies)

		// Get the set of commits within this repository that match an indexing policy
		commitMap, err := b.policyMatcher.CommitsDescribedByPolicy(ctx, repositoryID, policies, now)
		if err != nil {
			return errors.Wrap(err, "policies.CommitsDescribedByPolicy")
		}

		for commit, policyMatches := range commitMap {
			if len(policyMatches) == 0 {
				continue
			}

			uniqueCommits[commit] = struct{}{}

			// Attempt to queue an index if one does not exist for each of the matching commits
			if _, err := b.autoindexingSvc.QueueIndexes(ctx, repositoryID, commit, "", false, false); err != nil {
				if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
					continue
				}

				return errors.Wrap(err, "indexEnqueuer.QueueIndexes")
			}
		}

		if len(policies) == 0 || offset >= totalCount {
			return nil
		}
	}
}

func (b backgroundJob) NewOnDemandScheduler(interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		if !autoIndexingEnabled() {
			return nil
		}

		return b.autoindexingSvc.ProcessRepoRevs(ctx, batchSize)
	}))
}

func intp(v *int) int {
	if v == nil {
		return 0
	}

	return *v
}
