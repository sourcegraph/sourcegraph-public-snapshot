package local

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/internal/accesscontrol"
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
	defer noCache(ctx)
	return store.StorageFromContext(ctx).Create(ctx, opt)
}

// Remove deletes the named file or directory.
func (s *storage) Remove(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.Remove"); err != nil {
		return nil, err
	}
	defer noCache(ctx)
	return store.StorageFromContext(ctx).Remove(ctx, opt)
}

// RemoveAll deletes the named file or directory recursively.
func (s *storage) RemoveAll(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.RemoveAll"); err != nil {
		return nil, err
	}
	defer noCache(ctx)
	return store.StorageFromContext(ctx).RemoveAll(ctx, opt)
}

// Read reads from an existing file.
func (s *storage) Read(ctx context.Context, opt *sourcegraph.StorageReadOp) (*sourcegraph.StorageRead, error) {
	defer noCache(ctx)
	return store.StorageFromContext(ctx).Read(ctx, opt)
}

// Write writes to an existing file.
func (s *storage) Write(ctx context.Context, opt *sourcegraph.StorageWriteOp) (*sourcegraph.StorageWrite, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.Write"); err != nil {
		return nil, err
	}
	defer noCache(ctx)
	return store.StorageFromContext(ctx).Write(ctx, opt)
}

// Stat stats an existing file.
func (s *storage) Stat(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageStat, error) {
	defer noCache(ctx)
	return store.StorageFromContext(ctx).Stat(ctx, opt)
}

// ReadDir reads a directories contents.
func (s *storage) ReadDir(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageReadDir, error) {
	defer noCache(ctx)
	return store.StorageFromContext(ctx).ReadDir(ctx, opt)
}

// Close closes the named file or directory. You should always call Close once
// finished performing actions on a file.
func (s *storage) Close(ctx context.Context, opt *sourcegraph.StorageName) (*sourcegraph.StorageError, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.Close"); err != nil {
		return nil, err
	}
	defer noCache(ctx)
	return store.StorageFromContext(ctx).Close(ctx, opt)
}
