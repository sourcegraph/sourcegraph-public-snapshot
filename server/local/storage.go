package local

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

var Storage sourcegraph.StorageServer = &storage{}

var _ sourcegraph.StorageServer = (*storage)(nil)

type storage struct{}

// Create creates a new file with the given name.
func (s *storage) Create(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.Create"); err != nil {
		return nil, err
	}
	v1, v2 := store.StorageFromContext(ctx).Create(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// RemoveAll deletes the named file or directory recursively.
func (s *storage) RemoveAll(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.RemoveAll"); err != nil {
		return nil, err
	}
	v1, v2 := store.StorageFromContext(ctx).RemoveAll(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// Read reads from an existing file.
func (s *storage) Read(ctx context.Context, opt *sourcegraph.StorageReadOp) (*sourcegraph.StorageRead, error) {
	v1, v2 := store.StorageFromContext(ctx).Read(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// Write writes to an existing file.
func (s *storage) Write(ctx context.Context, opt *sourcegraph.StorageWriteOp) (*sourcegraph.StorageWrite, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.Write"); err != nil {
		return nil, err
	}
	v1, v2 := store.StorageFromContext(ctx).Write(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// Stat stats an existing file.
func (s *storage) Stat(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageStat, error) {
	v1, v2 := store.StorageFromContext(ctx).Stat(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// ReadDir reads a directories contents.
func (s *storage) ReadDir(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageReadDir, error) {
	v1, v2 := store.StorageFromContext(ctx).ReadDir(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// Close closes the named file or directory. You should always call Close once
// finished performing actions on a file.
func (s *storage) Close(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.Close"); err != nil {
		return nil, err
	}
	v1, v2 := store.StorageFromContext(ctx).Close(ctx, opt)
	noCache(ctx)
	return v1, v2
}
