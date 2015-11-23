package local

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

var Storage sourcegraph.StorageServer = &storage{}

var _ sourcegraph.StorageServer = (*storage)(nil)

type storage struct{}

// Get implements the sourcegraph.StorageServer interface.
func (s *storage) Get(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageValue, error) {
	v1, v2 := store.StorageFromContext(ctx).Get(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// Put implements the sourcegraph.StorageServer interface.
func (s *storage) Put(ctx context.Context, opt *sourcegraph.StoragePutOp) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.Put"); err != nil {
		return nil, err
	}
	v1, v2 := store.StorageFromContext(ctx).Put(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// Delete implements the sourcegraph.StorageServer interface.
func (s *storage) Delete(ctx context.Context, opt *sourcegraph.StorageKey) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Storage.Delete"); err != nil {
		return nil, err
	}
	v1, v2 := store.StorageFromContext(ctx).Delete(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// Exists implements the sourcegraph.StorageServer interface.
func (s *storage) Exists(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageExists, error) {
	v1, v2 := store.StorageFromContext(ctx).Exists(ctx, opt)
	noCache(ctx)
	return v1, v2
}

// List implements the sourcegraph.StorageServer interface.
func (s *storage) List(ctx context.Context, opt *sourcegraph.StorageKey) (*sourcegraph.StorageList, error) {
	v1, v2 := store.StorageFromContext(ctx).List(ctx, opt)
	noCache(ctx)
	return v1, v2
}
