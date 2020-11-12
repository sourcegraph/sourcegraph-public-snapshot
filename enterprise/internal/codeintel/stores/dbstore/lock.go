package dbstore

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"
	"github.com/segmentio/fasthash/fnv1"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// appLockKey is the namespace in which all advisory locks are taken.
var appLockKey = int(fnv1.HashString32("codeintel"))

// UnlockFunc unlocks the advisory lock taken by a successful call to Lock. If an error
// occurs during unlock, the error is added to the resulting error value.
type UnlockFunc func(err error) error

// Lock attempts to take an advisory lock on the given key. If successful, this method will
// return a true-valued flag along with a function that must be called to release the lock.
func (s *Store) Lock(ctx context.Context, key int, blocking bool) (locked bool, _ UnlockFunc, err error) {
	ctx, endObservation := s.operations.lock.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("key", key),
		log.Bool("blocking", blocking),
	}})
	defer endObservation(1, observation.Args{})

	if blocking {
		locked, err = s.lock(ctx, key)
	} else {
		locked, err = s.tryLock(ctx, key)
	}

	if err != nil || !locked {
		return false, nil, err
	}

	unlock := func(err error) error {
		if unlockErr := s.unlock(key); unlockErr != nil {
			err = multierror.Append(err, unlockErr)
		}

		return err
	}

	return true, unlock, nil
}

// lock blocks until an advisory lock is taken on the given key.
func (s *Store) lock(ctx context.Context, key int) (bool, error) {
	err := s.Store.Exec(ctx, sqlf.Sprintf(`SELECT pg_advisory_lock(%s, %s)`, appLockKey, key))
	if err != nil {
		return false, err
	}
	return true, nil
}

// tryLock attempts to tak ean advisory lock on the given key. Returns true on
// success and false on failure.
func (s *Store) tryLock(ctx context.Context, key int) (bool, error) {
	ok, _, err := basestore.ScanFirstBool(s.Store.Query(ctx, sqlf.Sprintf(`SELECT pg_try_advisory_lock(%s, %s)`, appLockKey, key)))
	if err != nil || !ok {
		return false, err
	}
	return true, nil
}

// unlock releases the advisory lock on the given key.
func (s *Store) unlock(key int) error {
	err := s.Store.Exec(context.Background(), sqlf.Sprintf(`SELECT pg_advisory_unlock(%s, %s)`, appLockKey, key))
	return err
}
