package service

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/object"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

// New returns a Service.
func New(
	observationCtx *observation.Context,
	store *store.Store,
	uploadStore object.Storage,
	newSearcher NewSearcher,
) *Service {
	logger := log.Scoped("searchjobs.Service")

	svc := &Service{
		logger:      logger,
		store:       store,
		uploadStore: uploadStore,
		newSearcher: newSearcher,
		operations:  newOperations(observationCtx),
	}

	return svc
}

type Service struct {
	logger      log.Logger
	store       *store.Store
	uploadStore object.Storage
	newSearcher NewSearcher
	operations  *operations
}

func opAttrs(attrs ...attribute.KeyValue) observation.Args {
	return observation.Args{Attrs: attrs}
}

type operations struct {
	createSearchJob          *observation.Operation
	getSearchJob             *observation.Operation
	deleteSearchJob          *observation.Operation
	listSearchJobs           *observation.Operation
	cancelSearchJob          *observation.Operation
	getAggregateRepoRevState *observation.Operation

	getSearchJobResultsWriterTo operationWithWriterTo
	getSearchJobLogsWriterTo    operationWithWriterTo
}

// operationWithWriterTo encodes our pattern around our CSV WriterTo were we
// have two steps that run adjacent to each other. First validating we can get
// the job, then we return a WriterTo which actually writes.
type operationWithWriterTo struct {
	get      *observation.Operation
	writerTo *observation.Operation
}

var (
	singletonOperations *operations
	operationsOnce      sync.Once
)

// newOperations generates a singleton of the operations struct.
//
// TODO: We should create one per observationCtx. This is a copy-pasta from
// the batches service, we should validate if we need to do this protection.
func newOperations(observationCtx *observation.Context) *operations {
	operationsOnce.Do(func() {
		m := metrics.NewREDMetrics(
			observationCtx.Registerer,
			"searchjobs_service",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)

		op := func(name string) *observation.Operation {
			return observationCtx.Operation(observation.Op{
				Name:              fmt.Sprintf("searchjobs.service.%s", name),
				MetricLabelValues: []string{name},
				Metrics:           m,
			})
		}

		singletonOperations = &operations{
			createSearchJob:          op("CreateSearchJob"),
			getSearchJob:             op("GetSearchJob"),
			deleteSearchJob:          op("DeleteSearchJob"),
			listSearchJobs:           op("ListSearchJobs"),
			cancelSearchJob:          op("CancelSearchJob"),
			getAggregateRepoRevState: op("GetAggregateRepoRevState"),

			getSearchJobResultsWriterTo: operationWithWriterTo{
				get:      op("GetSearchJobResultsWriterTo"),
				writerTo: op("GetSearchJobResultsWriterTo.WriteTo"),
			},
			getSearchJobLogsWriterTo: operationWithWriterTo{
				get:      op("GetSearchJobLogsWriterTo"),
				writerTo: op("GetSearchJobLogsWriterTo.WriteTo"),
			},
		}
	})
	return singletonOperations
}

func (s *Service) ValidateSearchJob(ctx context.Context, query string) error {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return errors.New("search jobs can only be validated by an authenticated user")
	}
	_, err := s.newSearcher.NewSearch(ctx, actor.UID, query)
	return err
}

func (s *Service) CreateSearchJob(ctx context.Context, query string) (_ *types.ExhaustiveSearchJob, err error) {
	ctx, _, endObservation := s.operations.createSearchJob.With(ctx, &err, opAttrs(
		attribute.String("query", query),
	))
	defer endObservation(1, observation.Args{})

	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("search jobs can only be created by an authenticated user")
	}

	// Validate query
	err = s.ValidateSearchJob(ctx, query)
	if err != nil {
		return nil, err
	}

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	// XXX(keegancsmith) this API for creating seems easy to mess up since the
	// ExhaustiveSearchJob type has lots of fields, but reading the store
	// implementation only two fields are read.
	jobID, err := tx.CreateExhaustiveSearchJob(ctx, types.ExhaustiveSearchJob{
		InitiatorID: actor.UID,
		Query:       query,
	})
	if err != nil {
		return nil, err
	}

	return tx.GetExhaustiveSearchJob(ctx, jobID)
}

func (s *Service) CancelSearchJob(ctx context.Context, id int64) (err error) {
	ctx, _, endObservation := s.operations.cancelSearchJob.With(ctx, &err, opAttrs(
		attribute.Int64("id", id),
	))
	defer endObservation(1, observation.Args{})

	tx, err := s.store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	_, err = tx.CancelSearchJob(ctx, id)
	return err
}

func (s *Service) GetSearchJob(ctx context.Context, id int64) (_ *types.ExhaustiveSearchJob, err error) {
	ctx, _, endObservation := s.operations.getSearchJob.With(ctx, &err, opAttrs(
		attribute.Int64("id", id),
	))
	defer endObservation(1, observation.Args{})

	return s.store.GetExhaustiveSearchJob(ctx, id)
}

func (s *Service) ListSearchJobs(ctx context.Context, args store.ListArgs) (jobs []*types.ExhaustiveSearchJob, err error) {
	ctx, _, endObservation := s.operations.listSearchJobs.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, opAttrs(
			attribute.Int("len", len(jobs)),
		))
	}()

	return s.store.ListExhaustiveSearchJobs(ctx, args)
}

// GetSearchJobLogsWriterTo returns a WriterTo which can be called once to
// write the logs for job id. Note: ctx is used by WriterTo.
//
// io.WriterTo is a specialization of an io.Reader. We expect callers of this
// function to want to write an http response, so we avoid an io.Pipe and
// instead pass a more direct use.
func (s *Service) GetSearchJobLogsWriterTo(parentCtx context.Context, id int64) (_ io.WriterTo, err error) {
	ctx, _, endObservation := s.operations.getSearchJobLogsWriterTo.get.With(parentCtx, &err, opAttrs(
		attribute.Int64("id", id)))
	defer endObservation(1, observation.Args{})

	// 🚨 SECURITY: only someone with access to the job may copy the blobs
	if err := s.store.UserHasAccess(ctx, id); err != nil {
		return nil, err
	}

	iter := s.getJobLogsIter(ctx, id)

	return writerToFunc(func(w io.Writer) (n int64, err error) {
		_, _, endObservation := s.operations.getSearchJobLogsWriterTo.writerTo.With(parentCtx, &err, opAttrs(
			attribute.Int64("id", id)))
		defer func() {
			endObservation(1, opAttrs(attribute.Int64("bytesWritten", n)))
		}()

		return writeSearchJobLogs(iter, w)
	}), nil
}

// getLogKey returns the key for the log that is stored in the blobstore.
func getLogKey(searchJobID int64) string {
	return fmt.Sprintf("log-%d.csv", searchJobID)
}

func (s *Service) UploadJobLogs(ctx context.Context, id int64, r io.Reader) (int64, error) {
	// 🚨 SECURITY: only someone with access to the job may upload the logs
	if err := s.store.UserHasAccess(ctx, id); err != nil {
		return 0, err
	}

	return s.uploadStore.Upload(ctx, getLogKey(id), r)
}

func (s *Service) GetJobLogs(ctx context.Context, id int64) (io.ReadCloser, error) {
	// 🚨 SECURITY: only someone with access to the job may download the logs
	if err := s.store.UserHasAccess(ctx, id); err != nil {
		return nil, err
	}

	return s.uploadStore.Get(ctx, getLogKey(id))
}

func (s *Service) DeleteJobLogs(ctx context.Context, id int64) error {
	// 🚨 SECURITY: only someone with access to the job may delete the logs
	if err := s.store.UserHasAccess(ctx, id); err != nil {
		return err
	}

	return s.uploadStore.Delete(ctx, getLogKey(id))
}

// JobLogsIterLimit is the number of lines the iterator will read from the
// database per page. Assuming 100 bytes per line, this will be ~1MB of memory
// per 10k repo-rev jobs.
var JobLogsIterLimit = 10_000

func (s *Service) getJobLogsIter(ctx context.Context, id int64) *iterator.Iterator[types.SearchJobLog] {
	var cursor int64
	limit := JobLogsIterLimit

	return iterator.New(func() ([]types.SearchJobLog, error) {
		if cursor == -1 {
			return nil, nil
		}

		opts := &store.GetJobLogsOpts{
			From:  cursor,
			Limit: limit + 1,
		}

		logs, err := s.store.GetJobLogs(ctx, id, opts)
		if err != nil {
			return nil, err
		}

		if len(logs) > limit {
			cursor = logs[len(logs)-1].ID
			logs = logs[:len(logs)-1]
		} else {
			cursor = -1
		}

		return logs, nil
	})
}

func formatOrNULL(t time.Time) string {
	if t.IsZero() {
		return "NULL"
	}

	return t.Format(time.RFC3339)
}

func getPrefix(id int64) string {
	return fmt.Sprintf("%d-", id)
}

func (s *Service) DeleteSearchJob(ctx context.Context, id int64) (err error) {
	ctx, _, endObservation := s.operations.deleteSearchJob.With(ctx, &err, opAttrs(
		attribute.Int64("id", id)))
	defer func() {
		endObservation(1, observation.Args{})
	}()

	// 🚨 SECURITY: only someone with access to the job may delete data and the db entries
	if err := s.store.UserHasAccess(ctx, id); err != nil {
		return err
	}

	iter, err := s.uploadStore.List(ctx, getPrefix(id))
	if err != nil {
		return err
	}
	for iter.Next() {
		key := iter.Current()
		err := s.uploadStore.Delete(ctx, key)
		// If we continued, we might end up with data in the upload store without
		// entries in the db to reference it.
		if err != nil {
			return errors.Wrapf(err, "deleting key %q", key)
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	// The log file is not guaranteed to exist, so we ignore the error here.
	_ = s.uploadStore.Delete(ctx, getLogKey(id))

	return s.store.DeleteExhaustiveSearchJob(ctx, id)
}

// GetSearchJobResultsWriterTo returns a WriterTo which can be called once to
// write all CSVs associated with a search job to the given writer for job id.
// Note: ctx is used by WriterTo.
//
// io.WriterTo is a specialization of an io.Reader. We expect callers of this
// function to want to write a http response, so we avoid an io.Pipe and
// instead pass a more direct use.
func (s *Service) GetSearchJobResultsWriterTo(parentCtx context.Context, id int64) (_ io.WriterTo, err error) {
	ctx, _, endObservation := s.operations.getSearchJobResultsWriterTo.get.With(parentCtx, &err, opAttrs(
		attribute.Int64("id", id)))
	defer endObservation(1, observation.Args{})

	// 🚨 SECURITY: only someone with access to the job may copy the blobs
	if err := s.store.UserHasAccess(ctx, id); err != nil {
		return nil, err
	}

	iter, err := s.uploadStore.List(ctx, getPrefix(id))
	if err != nil {
		return nil, err
	}

	return writerToFunc(func(w io.Writer) (n int64, err error) {
		ctx, _, endObservation := s.operations.getSearchJobResultsWriterTo.writerTo.With(parentCtx, &err, opAttrs(
			attribute.Int64("id", id)))
		defer func() {
			endObservation(1, opAttrs(attribute.Int64("bytesWritten", n)))
		}()

		return writeSearchJobJSON(ctx, iter, s.uploadStore, w)
	}), nil
}

// GetAggregateRepoRevState returns the map of state -> count for all repo
// revision jobs for the given job.
func (s *Service) GetAggregateRepoRevState(ctx context.Context, id int64) (_ *types.RepoRevJobStats, err error) {
	ctx, _, endObservation := s.operations.getAggregateRepoRevState.With(ctx, &err, opAttrs(
		attribute.Int64("id", id)))
	defer endObservation(1, observation.Args{})

	m, err := s.store.GetAggregateRepoRevState(ctx, id)
	if err != nil {
		return nil, err
	}

	stats := types.RepoRevJobStats{}
	for state, count := range m {
		switch types.JobState(state) {
		case types.JobStateCompleted:
			stats.Completed += int32(count)
		case types.JobStateFailed:
			stats.Failed += int32(count)
		case types.JobStateProcessing, types.JobStateErrored, types.JobStateQueued:
			stats.InProgress += int32(count)
		case types.JobStateCanceled:
			stats.InProgress = 0
		default:
			return nil, errors.Newf("unknown job state %q", state)
		}
	}

	stats.Total = stats.Completed + stats.Failed + stats.InProgress

	return &stats, nil
}

func writeSearchJobJSON(ctx context.Context, iter *iterator.Iterator[string], uploadStore object.Storage, w io.Writer) (int64, error) {
	// keep a single bufio.Reader so we can reuse its buffer.
	var br bufio.Reader

	writeKey := func(key string) (int64, error) {
		rc, err := uploadStore.Get(ctx, key)
		if err != nil {
			return 0, err
		}
		defer rc.Close()

		br.Reset(rc)

		return br.WriteTo(w)
	}

	var n int64
	for iter.Next() {
		key := iter.Current()
		m, err := writeKey(key)
		n += m
		if err != nil {
			return n, errors.Wrapf(err, "writing JSON for key %q", key)
		}
	}

	return n, iter.Err()
}

func writeSearchJobLogs(iter *iterator.Iterator[types.SearchJobLog], w io.Writer) (int64, error) {
	// For csv.NewWriter we have no way to track bytes written, so we wrap
	// w to find out. The implementation of csv writer uses a
	// bufio.NewWriter and avoids any uses of optimized interfaces like
	// WriterTo so this has no perf impact.
	writeCounter := &writeCounter{w: w}
	cw := csv.NewWriter(writeCounter)

	header := []string{
		"repository",
		"revision",
		"started_at",
		"finished_at",
		"status",
		"failure_message",
	}
	err := cw.Write(header)
	if err != nil {
		return writeCounter.n, err
	}

	for iter.Next() {
		job := iter.Current()
		err = cw.Write([]string{
			string(job.RepoName),
			job.Revision,
			formatOrNULL(job.StartedAt),
			formatOrNULL(job.FinishedAt),
			string(job.State),
			job.FailureMessage,
		})
		if err != nil {
			return writeCounter.n, err
		}
	}

	if err := iter.Err(); err != nil {
		return writeCounter.n, err
	}

	// Flush data before checking for any final write errors.
	cw.Flush()
	return writeCounter.n, cw.Error()
}

type writerToFunc func(w io.Writer) (int64, error)

func (f writerToFunc) WriteTo(w io.Writer) (int64, error) {
	return f(w)
}

// writeCounter wraps an io.Writer and keeps track of bytes written.
type writeCounter struct {
	w io.Writer
	// n is the number of bytes written to w
	n int64
}

func (c *writeCounter) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	c.n += int64(n)
	return
}
