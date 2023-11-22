package dbworker

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// storeShim converts a store.Store into a workerutil.Store.
type storeShim[T workerutil.Record] struct {
	store.Store[T]
}

var _ workerutil.Store[workerutil.Record] = &storeShim[workerutil.Record]{}

// newStoreShim wraps the given store in a shim.
func newStoreShim[T workerutil.Record](store store.Store[T]) workerutil.Store[T] {
	if store == nil {
		return nil
	}

	return &storeShim[T]{Store: store}
}

// QueuedCount calls into the inner store.
func (s *storeShim[T]) QueuedCount(ctx context.Context) (int, error) {
	return s.Store.QueuedCount(ctx, false)
}

// Dequeue calls into the inner store.
func (s *storeShim[T]) Dequeue(ctx context.Context, workerHostname string, extraArguments any) (ret T, _ bool, _ error) {
	conditions, err := convertArguments(extraArguments)
	if err != nil {
		return ret, false, err
	}

	return s.Store.Dequeue(ctx, workerHostname, conditions)
}

func (s *storeShim[T]) Heartbeat(ctx context.Context, ids []string) (knownIDs, cancelIDs []string, err error) {
	return s.Store.Heartbeat(ctx, ids, store.HeartbeatOptions{})
}

func (s *storeShim[T]) MarkComplete(ctx context.Context, rec T) (bool, error) {
	return s.Store.MarkComplete(ctx, rec.RecordID(), store.MarkFinalOptions{})
}

func (s *storeShim[T]) MarkFailed(ctx context.Context, rec T, failureMessage string) (bool, error) {
	return s.Store.MarkFailed(ctx, rec.RecordID(), failureMessage, store.MarkFinalOptions{})
}

func (s *storeShim[T]) MarkErrored(ctx context.Context, rec T, errorMessage string) (bool, error) {
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
