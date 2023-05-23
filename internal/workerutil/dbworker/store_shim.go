package dbworker

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// storeShim converts a store.Store into a workerutil.Store.
type storeShim[T workerutil.Record, U workerutil.Record] struct {
	store.Store[T]
}

var _ workerutil.Store[workerutil.Record, workerutil.IntRecord] = &storeShim[workerutil.Record, workerutil.IntRecord]{}

// newStoreShim wraps the given store in a shim.
func newStoreShim[T workerutil.Record, U workerutil.Record](store store.Store[T, U]) workerutil.Store[T, U] {
	if store == nil {
		return nil
	}

	return &storeShim[T, U]{Store: store}
}

// QueuedCount calls into the inner store.
func (s *storeShim[T, U]) QueuedCount(ctx context.Context) (int, error) {
	return s.Store.QueuedCount(ctx, false)
}

// Dequeue calls into the inner store.
func (s *storeShim[T, U]) Dequeue(ctx context.Context, workerHostname string, extraArguments any) (ret T, _ bool, _ error) {
	conditions, err := convertArguments(extraArguments)
	if err != nil {
		return ret, false, err
	}

	return s.Store.Dequeue(ctx, workerHostname, conditions)
}

func (s *storeShim[T, U]) Heartbeat(ctx context.Context, ids []U) (knownIDs, cancelIDs []U, err error) {
	intIDs := make([]int, len(ids))
	for i, id := range ids {
		intIDs[i] = id.RecordID()
	}
	k, c, err := s.Store.Heartbeat(ctx, intIDs, store.HeartbeatOptions{})
	return intSliceToIntRecordSlice(k), intSliceToIntRecordSlice(c), err
}

func (s *storeShim[T, U]) MarkComplete(ctx context.Context, rec T) (bool, error) {
	return s.Store.MarkComplete(ctx, rec.RecordID(), store.MarkFinalOptions{})
}

func (s *storeShim[T, U]) MarkFailed(ctx context.Context, rec T, failureMessage string) (bool, error) {
	return s.Store.MarkFailed(ctx, rec.RecordID(), failureMessage, store.MarkFinalOptions{})
}

func (s *storeShim[T, U]) MarkErrored(ctx context.Context, rec T, errorMessage string) (bool, error) {
	return s.Store.MarkErrored(ctx, rec.RecordID(), errorMessage, store.MarkFinalOptions{})
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
