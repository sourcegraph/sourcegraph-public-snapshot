package syntactic_indexing

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.opentelemetry.io/otel/attribute"
)

type IndexEnqueuer interface {
	QueueIndexingJobs(ctx context.Context, repositoryId api.RepoID, commitId api.CommitID, options EnqueueOptions) (_ []jobstore.SyntacticIndexingJob, err error)
}

type EnqueueOptions struct {
	// setting force=true will schedule the job for a
	// given pair of (repo, commit) even if it already exists in the queue
	force bool
}

type indexEnqueuerImpl struct {
	jobStore            jobstore.SyntacticIndexingJobStore
	repoSchedulingStore reposcheduler.RepositorySchedulingStore
	repoStore           database.RepoStore
	operations          *operations
}

var _ IndexEnqueuer = &indexEnqueuerImpl{}

func NewIndexEnqueuer(
	observationCtx *observation.Context,
	jobStore jobstore.SyntacticIndexingJobStore,
	store reposcheduler.RepositorySchedulingStore,
	repoStore database.RepoStore,
) IndexEnqueuer {
	return &indexEnqueuerImpl{
		repoSchedulingStore: store,
		repoStore:           repoStore,
		jobStore:            jobStore,
		operations:          newOperations(observationCtx),
	}
}

type operations struct {
	queueIndexingJobs          *observation.Operation
	indexingJobsSkippedCounter prometheus.Counter
}

var (
	indexingJobsSkippedCounterMemo = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (prometheus.Counter, error) {
		indexesJobsSkippedCounter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "src_codeintel_dbstore_syntactic_indexing_jobs_skipped",
			Help: "The number of codeintel syntactic indexing jobs skipped because they were already queued up",
		})
		r.MustRegister(indexesJobsSkippedCounter)
		return indexesJobsSkippedCounter, nil
	})

	m = new(metrics.SingletonREDMetrics)
)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_syntactic_indexing_enqueuer",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.syntactic_indexing.enqueuer.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	indexingJobsSkippedCounter, _ := indexingJobsSkippedCounterMemo.Init(observationCtx.Registerer)

	return &operations{
		queueIndexingJobs:          op("QueueIndexingJobs"),
		indexingJobsSkippedCounter: indexingJobsSkippedCounter,
	}
}

// QueueIndexingJobs schedules a syntactic indexing job for a given repositoryID at given revision.
// If options.force = true, then the job will be scheduled even if the same one already exists in the queue.
// This method will return an array of jobs that were actually successfully scheduled.
// The result can be nil iff the same job is already queued AND options.force is false.
func (s *indexEnqueuerImpl) QueueIndexingJobs(ctx context.Context, repositoryID api.RepoID, commitID api.CommitID, options EnqueueOptions) (_ []jobstore.SyntacticIndexingJob, err error) {
	ctx, _, endObservation := s.operations.queueIndexingJobs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", int(repositoryID)),
		attribute.String("commitID", string(commitID)),
	}})
	defer endObservation(1, observation.Args{})

	shouldInsert := true
	if !options.force {
		isQueued, err := s.jobStore.IsQueued(ctx, repositoryID, commitID)
		if err != nil {
			return nil, errors.Wrap(err, "dbstore.IsQueued")
		}
		if isQueued {
			s.operations.indexingJobsSkippedCounter.Add(float64(1))
		}
		shouldInsert = !isQueued
	}
	if shouldInsert {
		return s.jobStore.InsertIndexingJobs(ctx, []jobstore.SyntacticIndexingJob{{
			State:        jobstore.Queued,
			Commit:       commitID,
			RepositoryID: repositoryID,
		}})
	}
	return nil, nil
}
