package dbworker

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// storeShim converts a store.Store into a workerutil.Store.
type storeShim struct {
	store.Store
}

var _ workerutil.Store = &storeShim{}

// newStoreShim wraps the given store in a shim.
func newStoreShim(store store.Store) workerutil.Store {
	if store == nil {
		return nil
	}

	return &storeShim{Store: store}
}

// QueuedCount calls into the inner store.
func (s *storeShim) QueuedCount(ctx context.Context, extraArguments any) (int, error) {
	conditions, err := convertArguments(extraArguments)
	if err != nil {
		return 0, err
	}

	return s.Store.QueuedCount(ctx, false, conditions)
}

// Dequeue calls into the inner store.
func (s *storeShim) Dequeue(ctx context.Context, workerHostname string, extraArguments any) (workerutil.Record, bool, error) {
	conditions, err := convertArguments(extraArguments)
	if err != nil {
		return nil, false, err
	}

	return s.Store.Dequeue(ctx, workerHostname, conditions)
}

func (s *storeShim) Heartbeat(ctx context.Context, ids []int) (knownIDs []int, err error) {
	return s.Store.Heartbeat(ctx, ids, store.HeartbeatOptions{})
}

func (s *storeShim) AddExecutionLogEntry(ctx context.Context, id int, entry workerutil.ExecutionLogEntry) (entryID int, err error) {
	return s.Store.AddExecutionLogEntry(ctx, id, entry, store.ExecutionLogEntryOptions{})
}

func (s *storeShim) UpdateExecutionLogEntry(ctx context.Context, recordID, entryID int, entry workerutil.ExecutionLogEntry) error {
	return s.Store.UpdateExecutionLogEntry(ctx, recordID, entryID, entry, store.ExecutionLogEntryOptions{})
}

func (s *storeShim) MarkComplete(ctx context.Context, id int) (bool, error) {
	return s.Store.MarkComplete(ctx, id, store.MarkFinalOptions{})
}

func (s *storeShim) MarkFailed(ctx context.Context, id int, failureMessage string) (bool, error) {
	return s.Store.MarkFailed(ctx, id, failureMessage, store.MarkFinalOptions{})
}

func (s *storeShim) MarkErrored(ctx context.Context, id int, errorMessage string) (bool, error) {
	return s.Store.MarkErrored(ctx, id, errorMessage, store.MarkFinalOptions{})
}

// ErrNotConditions occurs when a PreDequeue handler returns non-sql query extra arguments.
var ErrNotConditions = errors.New("expected slice of *sqlf.Query values")

// convertArguments converts the given interface value into a slice of *sqlf.Query values.
func convertArguments(v any) ([]*sqlf.Query, error) {
	if v == nil {
		return nil, nil
	}

	if conditions, ok := v.([]*sqlf.Query); ok {
		return conditions, nil
	}

	return nil, ErrNotConditions
}
