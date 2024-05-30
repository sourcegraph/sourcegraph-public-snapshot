package syntactic_indexing

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/reposcheduler"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/syntactic_indexing/jobstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.opentelemetry.io/otel/attribute"
)

type IndexEnqueuer interface {
	QueueIndexingJobs(ctx context.Context, repositoryId int, rev string, options EnqueueOptions) (_ []jobstore.SyntacticIndexingJob, err error)
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
	gitserverClient     gitserver.Client
	operations          *operations
}

var _ IndexEnqueuer = &indexEnqueuerImpl{}

func NewIndexEnqueuer(
	observationCtx *observation.Context,
	jobStore jobstore.SyntacticIndexingJobStore,
	store reposcheduler.RepositorySchedulingStore,
	repoStore database.RepoStore,
	gitserverClient gitserver.Client,
) IndexEnqueuer {
	return &indexEnqueuerImpl{
		repoSchedulingStore: store,
		repoStore:           repoStore,
		gitserverClient:     gitserverClient,
		jobStore:            jobStore,
		operations:          newOperations(observationCtx),
	}
}

type operations struct {
	queueIndexingJobs *observation.Operation
}

var (
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

	return &operations{
		queueIndexingJobs: op("QueueIndexingJobs"),
	}
}

// QueueIndexingJobs schedules a syntactic indexing job for a given repositoryID at given revision.
// If options.force = true, then the job will be scheduled even if the same one already exists in the queue.
// This method will return an array of jobs that were actually successfully scheduled.
// The result can be nil iff the same job is already queued AND options.force is false.
func (s *indexEnqueuerImpl) QueueIndexingJobs(ctx context.Context, repositoryID int, rev string, options EnqueueOptions) (_ []jobstore.SyntacticIndexingJob, err error) {
	ctx, trace, endObservation := s.operations.queueIndexingJobs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("repositoryID", repositoryID),
		attribute.String("rev", rev),
	}})
	defer endObservation(1, observation.Args{})

	repo, err := s.repoStore.Get(ctx, api.RepoID(repositoryID))
	if err != nil {
		return nil, err
	}

	commitID, err := s.gitserverClient.ResolveRevision(ctx, repo.Name, rev, gitserver.ResolveRevisionOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "gitserver.ResolveRevision")
	}
	trace.AddEvent("ResolveRevision", attribute.String("commit", string(commitID)))

	return s.queueIndexForRepositoryAndCommit(ctx, repositoryID, commitID, options)
}

func (s *indexEnqueuerImpl) queueIndexForRepositoryAndCommit(ctx context.Context, repositoryID int, commitID api.CommitID, options EnqueueOptions) ([]jobstore.SyntacticIndexingJob, error) {
	commit := string(commitID)
	shouldInsert := true
	if !options.force {
		isQueued, err := s.jobStore.IsQueued(ctx, repositoryID, commit)
		if err != nil {
			return nil, errors.Wrap(err, "dbstore.IsQueued")
		}
		shouldInsert = !isQueued
	}
	if shouldInsert {
		return s.jobStore.InsertIndexes(ctx, []jobstore.SyntacticIndexingJob{{
			State:        jobstore.Queued,
			Commit:       commit,
			RepositoryID: repositoryID,
		}})
	}
	return nil, nil
}
