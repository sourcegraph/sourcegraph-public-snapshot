package store

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// QueueShim wraps QueueStore to implement workerutil.Store.
type QueueShim struct {
	Name  string
	Store QueueStore
}

type QueueStore interface {
	Dequeue(ctx context.Context, queueName string, payload *executor.Job) (bool, error)
	AddExecutionLogEntry(ctx context.Context, queueName string, jobID int, entry workerutil.ExecutionLogEntry) (int, error)
	UpdateExecutionLogEntry(ctx context.Context, queueName string, jobID, entryID int, entry workerutil.ExecutionLogEntry) error
	MarkComplete(ctx context.Context, queueName string, jobID int) error
	MarkErrored(ctx context.Context, queueName string, jobID int, errorMessage string) error
	MarkFailed(ctx context.Context, queueName string, jobID int, errorMessage string) error
	Heartbeat(ctx context.Context, queueName string, jobIDs []int) (knownIDs []int, err error)
	CanceledJobs(ctx context.Context, queueName string, knownIDs []int) (canceledIDs []int, err error)
}

// Compile time validation.
var _ workerutil.Store = &QueueShim{}

func (s *QueueShim) QueuedCount(ctx context.Context) (int, error) {
	return 0, errors.New("unimplemented")
}

func (s *QueueShim) Dequeue(ctx context.Context, workerHostname string, extraArguments any) (workerutil.Record, bool, error) {
	var job executor.Job
	dequeued, err := s.Store.Dequeue(ctx, s.Name, &job)
	if err != nil {
		return nil, false, err
	}

	return job, dequeued, nil
}

func (s *QueueShim) Heartbeat(ctx context.Context, ids []int) (knownIDs []int, err error) {
	return s.Store.Heartbeat(ctx, s.Name, ids)
}

func (s *QueueShim) AddExecutionLogEntry(ctx context.Context, id int, entry workerutil.ExecutionLogEntry) (int, error) {
	return s.Store.AddExecutionLogEntry(ctx, s.Name, id, entry)
}

func (s *QueueShim) UpdateExecutionLogEntry(ctx context.Context, jobID, entryID int, entry workerutil.ExecutionLogEntry) error {
	return s.Store.UpdateExecutionLogEntry(ctx, s.Name, jobID, entryID, entry)
}

func (s *QueueShim) MarkComplete(ctx context.Context, id int) (bool, error) {
	return true, s.Store.MarkComplete(ctx, s.Name, id)
}

func (s *QueueShim) MarkErrored(ctx context.Context, id int, errorMessage string) (bool, error) {
	return true, s.Store.MarkErrored(ctx, s.Name, id, errorMessage)
}

func (s *QueueShim) MarkFailed(ctx context.Context, id int, errorMessage string) (bool, error) {
	return true, s.Store.MarkFailed(ctx, s.Name, id, errorMessage)
}

func (s *QueueShim) CanceledJobs(ctx context.Context, knownIDs []int) ([]int, error) {
	return s.Store.CanceledJobs(ctx, s.Name, knownIDs)
}

// FilesStore handles interactions with the file store.
type FilesStore interface {
	// Exists determines if the file exists.
	Exists(ctx context.Context, bucket string, key string) (bool, error)
	// Get retrieves the file.
	Get(ctx context.Context, bucket string, key string) (io.ReadCloser, error)
}
