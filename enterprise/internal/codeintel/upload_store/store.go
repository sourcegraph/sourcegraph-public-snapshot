package uploadstore

import (
	"context"
	"fmt"
	"io"
)

// Store is an expiring key/value store backed by a managed blob store.
type Store interface {
	// Init ensures that the underlying target bucket exists and has the expected ACL
	// and lifecycle configuration.
	Init(ctx context.Context) error

	// Get returns a reader that streams the content of the object at the given key.
	// If a positive skipBytes is supplied, the reader will begin reading at that byte
	// offset.
	Get(ctx context.Context, key string, skipBytes int64) (io.ReadCloser, error)

	// Upload writes the content in the given reader to the object at the given key.
	Upload(ctx context.Context, key string, r io.Reader) (int64, error)

	// Compose will concatenate the given source objects together and write to the given
	// destination object. The source objects will be removed if the composed write is
	// successful.
	Compose(ctx context.Context, destination string, sources ...string) (int64, error)

	// Delete removes the content at the given key.
	Delete(ctx context.Context, key string) error
}

var storeConstructors = map[string]func(ctx context.Context, config *Config) (Store, error){
	"S3":  newS3FromConfig,
	"GCS": newGCSFromConfig,
}

// Create initialize a new store from the given configuration.
func Create(ctx context.Context, config *Config) (Store, error) {
	newStore, ok := storeConstructors[config.Backend]
	if !ok {
		return nil, fmt.Errorf("unknown upload store backend '%s'", config.Backend)
	}

	store, err := newStore(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := store.Init(ctx); err != nil {
		return nil, err
	}

	return store, err
}
