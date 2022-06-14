package locker

import (
	"context"
	"math"

	"github.com/keegancsmith/sqlf"
	"github.com/segmentio/fasthash/fnv1"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// StringKey returns an int32 key based on s that can be used in Locker methods.
func StringKey(s string) int32 {
	return int32(fnv1.HashString32(s) % math.MaxInt32)
}

// Locker is a wrapper around a base store with methods that control advisory locks.
// A locker should be used when work needs to be coordinated with other remote services.
//
// For example, an advisory lock can be taken around an expensive calculation related to
// a particular repository to ensure that no other service is performing the same task.
type Locker[T schemas.Any] struct {
	*basestore.Store[schemas.Any]
	namespace int32
}

// NewWith creates a new Locker with the given namespace and ShareableStore
func NewWith[T schemas.Any](other basestore.ShareableStore[T], namespace string) *Locker[T] {
	return &Locker[T]{
		Store:     basestore.NewWithHandle(other.Handle()),
		namespace: StringKey(namespace),
	}
}

func (l *Locker[T]) With(other basestore.ShareableStore[schemas.Any]) *Locker {
	return &Locker{
		Store:     l.Store.With(other),
		namespace: l.namespace,
	}
}

func (l *Locker[T]) Transact(ctx context.Context) (*Locker[T], error) {
	txBase, err := l.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Locker[T]{
		Store:     txBase,
		namespace: l.namespace,
	}, nil
}

// UnlockFunc unlocks the advisory lock taken by a successful call to Lock. If an error
// occurs during unlock, the error is added to the resulting error value.
type UnlockFunc func(error) error

// ErrTransaction occurs when Lock is called inside of a transaction.
var ErrTransaction = errors.New("locker: in a transaction")

// Lock creates a transactional store and calls its Lock method. This method expects that
// the locker is outside of a transaction. The transaction's lifetime is linked to the lock,
// so the internal locker will commit or rollback for the lock to be released.
func (l *Locker[T]) Lock(ctx context.Context, key int32, blocking bool) (locked bool, _ UnlockFunc, err error) {
	if l.InTransaction() {
		return false, nil, ErrTransaction
	}

	tx, err := l.Transact(ctx)
	if err != nil {
		return false, nil, err
	}
	defer func() {
		if !locked {
			// Catch failure cases
			err = tx.Done(err)
		}
	}()

	locked, err = tx.LockInTransaction(ctx, key, blocking)
	if err != nil || !locked {
		return false, nil, err
	}

	return true, tx.Done, nil
}

// ErrNoTransaction occurs when LockInTransaction is called outside of a transaction.
var ErrNoTransaction = errors.New("locker: not in a transaction")

// LockInTransaction attempts to take an advisory lock on the given key. If successful, this method
// will return a true-valued flag. This method assumes that the locker is currently in a transaction
// and will return an error if not. The lock is released when the transaction is committed or rolled back.
func (l *Locker) LockInTransaction(
	ctx context.Context,
	key int32,
	blocking bool,
) (locked bool, err error) {
	if !l.InTransaction() {
		return false, ErrNoTransaction
	}

	if blocking {
		locked, err = l.selectAdvisoryLock(ctx, key)
	} else {
		locked, err = l.selectTryAdvisoryLock(ctx, key)
	}

	if err != nil || !locked {
		return false, err
	}

	return true, nil
}

// selectAdvisoryLock blocks until an advisory lock is taken on the given key.
func (l *Locker) selectAdvisoryLock(ctx context.Context, key int32) (bool, error) {
	err := l.Store.Exec(ctx, sqlf.Sprintf(selectAdvisoryLockQuery, l.namespace, key))
	if err != nil {
		return false, err
	}
	return true, nil
}

const selectAdvisoryLockQuery = `
-- source: internal/database/locker/locker.go:selectAdvisoryLock
SELECT pg_advisory_xact_lock(%s, %s)
`

// selectTryAdvisoryLock attempts to take an advisory lock on the given key. Returns true
// on success and false on failure.
func (l *Locker) selectTryAdvisoryLock(ctx context.Context, key int32) (bool, error) {
	ok, _, err := basestore.ScanFirstBool(
		l.Store.Query(ctx, sqlf.Sprintf(selectTryAdvisoryLockQuery, l.namespace, key)),
	)
	if err != nil || !ok {
		return false, err
	}

	return true, nil
}

const selectTryAdvisoryLockQuery = `
-- source: internal/database/locker/locker.go:selectTryAdvisoryLock
SELECT pg_try_advisory_xact_lock(%s, %s)
`
