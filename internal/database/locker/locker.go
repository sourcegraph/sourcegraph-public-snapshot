package locker

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/segmentio/fasthash/fnv1"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// Locker is a wrapper around a base store with methods that control advisory locks.
// A locker should be used when work needs to be coordinated with other remote services.
//
// For example, an advisory lock can be taken around an expensive calculation related to
// a particular repository to ensure that no other service is performing the same task.
type Locker struct {
	*basestore.Store
	namespace int32
}

// NewWithDB creates a new Locker with the given namespace.
func NewWithDB(db dbutil.DB, namespace string) *Locker {
	return &Locker{
		Store:     basestore.NewWithDB(db, sql.TxOptions{}),
		namespace: int32(fnv1.HashString32(namespace)),
	}
}

func (l *Locker) With(other basestore.ShareableStore) *Locker {
	return &Locker{
		Store:     l.Store.With(other),
		namespace: l.namespace,
	}
}

func (l *Locker) Transact(ctx context.Context) (*Locker, error) {
	txBase, err := l.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &Locker{
		Store:     txBase,
		namespace: l.namespace,
	}, nil
}

// UnlockFunc unlocks the advisory lock taken by a successful call to Lock. If an error
// occurs during unlock, the error is added to the resulting error value.
type UnlockFunc func(err error) error

// Lock attempts to take an advisory lock on the given key. If successful, this method will
// return a true-valued flag along with a function that must be called to release the lock.
func (l *Locker) Lock(ctx context.Context, key int, blocking bool) (locked bool, _ UnlockFunc, err error) {
	if blocking {
		locked, err = l.lock(ctx, key)
	} else {
		locked, err = l.tryLock(ctx, key)
	}

	if err != nil || !locked {
		return false, nil, err
	}

	unlock := func(err error) error {
		if unlockErr := l.unlock(key); unlockErr != nil {
			err = multierror.Append(err, unlockErr)
		}

		return err
	}

	return true, unlock, nil
}

// lock blocks until an advisory lock is taken on the given key.
func (l *Locker) lock(ctx context.Context, key int) (bool, error) {
	err := l.Store.Exec(ctx, sqlf.Sprintf(lockQuery, l.namespace, key))
	if err != nil {
		return false, err
	}
	return true, nil
}

const lockQuery = `
-- source: internal/database/locker/locker.go:lock
SELECT pg_advisory_lock(%s, %s)
`

// tryLock attempts to take an advisory lock on the given key. Returns true on
// success and false on failure.
func (l *Locker) tryLock(ctx context.Context, key int) (bool, error) {
	ok, _, err := basestore.ScanFirstBool(l.Store.Query(ctx, sqlf.Sprintf(tryLockQuery, l.namespace, key)))
	if err != nil || !ok {
		return false, err
	}

	return true, nil
}

const tryLockQuery = `
-- source: internal/database/locker/locker.go:tryLock
SELECT pg_try_advisory_lock(%s, %s)
`

// unlock releases the advisory lock on the given key.
func (l *Locker) unlock(key int) error {
	err := l.Store.Exec(context.Background(), sqlf.Sprintf(unlockQuery, l.namespace, key))
	return err
}

const unlockQuery = `
-- source: internal/database/locker/locker.go:unlock
SELECT pg_advisory_unlock(%s, %s)
`
