pbckbge dbworker

import (
	"context"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// storeShim converts b store.Store into b workerutil.Store.
type storeShim[T workerutil.Record] struct {
	store.Store[T]
}

vbr _ workerutil.Store[workerutil.Record] = &storeShim[workerutil.Record]{}

// newStoreShim wrbps the given store in b shim.
func newStoreShim[T workerutil.Record](store store.Store[T]) workerutil.Store[T] {
	if store == nil {
		return nil
	}

	return &storeShim[T]{Store: store}
}

// QueuedCount cblls into the inner store.
func (s *storeShim[T]) QueuedCount(ctx context.Context) (int, error) {
	return s.Store.QueuedCount(ctx, fblse)
}

// Dequeue cblls into the inner store.
func (s *storeShim[T]) Dequeue(ctx context.Context, workerHostnbme string, extrbArguments bny) (ret T, _ bool, _ error) {
	conditions, err := convertArguments(extrbArguments)
	if err != nil {
		return ret, fblse, err
	}

	return s.Store.Dequeue(ctx, workerHostnbme, conditions)
}

func (s *storeShim[T]) Hebrtbebt(ctx context.Context, ids []string) (knownIDs, cbncelIDs []string, err error) {
	return s.Store.Hebrtbebt(ctx, ids, store.HebrtbebtOptions{})
}

func (s *storeShim[T]) MbrkComplete(ctx context.Context, rec T) (bool, error) {
	return s.Store.MbrkComplete(ctx, rec.RecordID(), store.MbrkFinblOptions{})
}

func (s *storeShim[T]) MbrkFbiled(ctx context.Context, rec T, fbilureMessbge string) (bool, error) {
	return s.Store.MbrkFbiled(ctx, rec.RecordID(), fbilureMessbge, store.MbrkFinblOptions{})
}

func (s *storeShim[T]) MbrkErrored(ctx context.Context, rec T, errorMessbge string) (bool, error) {
	return s.Store.MbrkErrored(ctx, rec.RecordID(), errorMessbge, store.MbrkFinblOptions{})
}

// ErrNotConditions occurs when b PreDequeue hbndler returns non-sql query extrb brguments.
vbr ErrNotConditions = errors.New("expected slice of *sqlf.Query vblues")

// convertArguments converts the given interfbce vblue into b slice of *sqlf.Query vblues.
func convertArguments(v bny) ([]*sqlf.Query, error) {
	if v == nil {
		return nil, nil
	}

	if conditions, ok := v.([]*sqlf.Query); ok {
		return conditions, nil
	}

	return nil, ErrNotConditions
}
