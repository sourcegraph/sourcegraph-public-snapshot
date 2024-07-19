// Package object provides an interface to object storage that abstracts over
// S3, GCS and blobstore.
package object

import (
	"context"
	"io"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

// Storage represents the API for some managed object storage.
//
// We use 'Storage' instead of 'Store' since 'Store' commonly refers to
// interfaces backed by database tables in the rest of the codebase.
type Storage interface {
	// Init ensures that the underlying target bucket exists and has the expected ACL
	// and lifecycle configuration.
	Init(ctx context.Context) error

	// Get returns a reader that streams the content of the object at the given key.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Upload writes the content in the given reader to the object at the given key.
	Upload(ctx context.Context, key string, r io.Reader) (int64, error)

	// Compose will concatenate the given source objects together and write to the given
	// destination object. The source objects will be removed if the composed write is
	// successful.
	Compose(ctx context.Context, destination string, sources ...string) (int64, error)

	// Delete removes the content at the given key.
	Delete(ctx context.Context, key string) error

	// ExpireObjects iterates all objects with the given prefix and deletes them when
	// the age of the object exceeds the given max age.
	ExpireObjects(ctx context.Context, prefix string, maxAge time.Duration) error

	// List returns an iterator over all keys with the given prefix.
	List(ctx context.Context, prefix string) (*iterator.Iterator[string], error)
}

var storeConstructors = map[string]func(ctx context.Context, config StorageConfig, operations *Operations) (Storage, error){
	"s3":        newS3FromConfig,
	"blobstore": newS3FromConfig,
	"gcs":       newGCSFromConfig,
}

// CreateLazyStorage initialize a new store from the given configuration that is initialized
// on it first method call. If initialization fails, all methods calls will return a
// the initialization error.
func CreateLazyStorage(ctx context.Context, config StorageConfig, ops *Operations) (Storage, error) {
	store, err := create(ctx, config, ops)
	if err != nil {
		return nil, err
	}

	return newLazyStore(store), nil
}

// create creates but does not initialize a new store from the given configuration.
func create(ctx context.Context, config StorageConfig, ops *Operations) (Storage, error) {
	newStore, ok := storeConstructors[config.Backend]
	if !ok {
		return nil, errors.Errorf("unknown upload store backend '%s'", config.Backend)
	}

	store, err := newStore(ctx, normalizeConfig(config), ops)
	if err != nil {
		return nil, err
	}

	return store, nil
}
