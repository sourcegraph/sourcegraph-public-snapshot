package store

import (
	"context"
	"io"
)

// FilesStore handles interactions with the file store.
type FilesStore interface {
	// Exists determines if the file exists.
	Exists(ctx context.Context, bucket string, key string) (bool, error)
	// Get retrieves the file.
	Get(ctx context.Context, bucket string, key string) (io.ReadCloser, error)
}
