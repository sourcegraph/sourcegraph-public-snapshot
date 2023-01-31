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
// We do not directly implement import that interface since that introduces
// complications around dependency graphs.
type DBStore interface {
	// Get returns the value for (namespace, key). ok is false if the
	// (namespace, key) has not been set.
	Get(ctx context.Context, namespace, key string) (value []byte, ok bool, err error)
	// Set will upsert value for (namespace, key).
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
// this is called all KeyValue operations fagainst a DB backed KeyValue will
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

// DBKeyValue returns a KeyValue with namespace. Namespaces allow us to have
// distinct KeyValue stores, but still use the same underlying DBStore
// storage.
func DBKeyValue(namespace string) KeyValue {
	var mu sync.Mutex
	store := func(ctx context.Context, key string, f NaiveUpdater) error {
		mu.Lock()
		defer mu.Unlock()

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
