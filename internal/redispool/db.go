package redispool

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DBStore is the methods needed by DBKeyValue to implement the core of
// KeyValue. See database.RedisKeyValueStore for the implementation of this
// interface.
//
// We do not directly import that interface since that introduces
// complications around dependency graphs.
//
// Note: DBKeyValue uses a coarse global mutex for all transactions on-top of
// whatever transaction DBStoreTransact provides. The intention of these
// interfaces is to be used in a single process application (like Sourcegraph
// App). We would need to change the design of NaiveKeyValueStore to allow for
// retries to smoothly avoid global mutexes.
type DBStore interface {
	// Get returns the value for (namespace, key). ok is false if the
	// (namespace, key) has not been set.
	//
	// Note: We recommend using "SELECT ... FOR UPDATE" since this call is
	// often followed by Set in the same transaction.
	Get(ctx context.Context, namespace, key string) (value []byte, ok bool, err error)
	// Set will upsert value for (namespace, key). If value is nil it should
	// be persisted as an empty byte slice.
	Set(ctx context.Context, namespace, key string, value []byte) (err error)
	// Delete will remove (namespace, key). If (namespace, key) is not in the
	// store, the delete is a noop.
	Delete(ctx context.Context, namespace, key string) (err error)
}

// DBStoreTransact is a function which is like the WithTransact which will run
// f inside of a transaction. f is a function which will read/update a
// DBStore.
type DBStoreTransact func(ctx context.Context, f func(DBStore) error) error

var dbStoreTransact atomic.Value

// DBRegisterStore registers our database with the redispool package. Until
// this is called all KeyValue operations against a DB backed KeyValue will
// fail with an error. As such this function should be called early on (as
// soon as we have a useable DB connection).
//
// An error will be returned if this function is called more than once.
func DBRegisterStore(transact DBStoreTransact) error {
	ok := dbStoreTransact.CompareAndSwap(nil, transact)
	if !ok {
		return errors.New("redispool.DBRegisterStore has already been called")
	}
	return nil
}

// dbMu protects _all_ possible interactions with the database in DBKeyValue.
// This is to avoid concurrent get/sets on the same key resulting in one of
// the sets failing due to serializability.
var dbMu sync.Mutex

// DBKeyValue returns a KeyValue with namespace. Namespaces allow us to have
// distinct KeyValue stores, but still use the same underlying DBStore
// storage.
//
// Note: This is designed for use in a single process application like
// Cody App. All transactions are additionally protected by a global
// mutex to avoid the need to handle database serializability errors.
func DBKeyValue(namespace string) KeyValue {
	store := func(ctx context.Context, key string, f NaiveUpdater) error {
		dbMu.Lock()
		defer dbMu.Unlock()

		transact := dbStoreTransact.Load()
		if transact == nil {
			return errors.New("redispool.DBRegisterStore has not been called")
		}

		return transact.(DBStoreTransact)(ctx, func(store DBStore) error {
			beforeStr, found, err := store.Get(ctx, namespace, key)
			if err != nil {
				return errors.Wrapf(err, "redispool.DBKeyValue failed to get %q in namespace %q", key, namespace)
			}

			before := NaiveValue(beforeStr)
			after, remove := f(before, found)
			if remove {
				if found {
					if err := store.Delete(ctx, namespace, key); err != nil {
						return errors.Wrapf(err, "redispool.DBKeyValue failed to delete %q in namespace %q", key, namespace)
					}
				}
			} else if before != after {
				if err := store.Set(ctx, namespace, key, []byte(after)); err != nil {
					return errors.Wrapf(err, "redispool.DBKeyValue failed to set %q in namespace %q", key, namespace)
				}
			}
			return nil
		})
	}

	return FromNaiveKeyValueStore(store)
}
