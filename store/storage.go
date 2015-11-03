package store

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type Storage interface {
	// Create creates a new file with the given name.
	Create(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error)

	// RemoveAll deletes the named file or directory recursively.
	RemoveAll(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error)

	// Read reads from an existing file.
	Read(ctx context.Context, opt *sourcegraph.StorageReadOp) (*sourcegraph.StorageRead, error)

	// Write writes to an existing file.
	Write(ctx context.Context, opt *sourcegraph.StorageWriteOp) (*sourcegraph.StorageWrite, error)

	// Stat stats an existing file.
	Stat(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageStat, error)

	// ReadDir reads a directories contents.
	ReadDir(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageReadDir, error)

	// Close closes the named file or directory. You should always call Close once
	// finished performing actions on a file.
	Close(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error)
}
