package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

var _ graphqlbackend.BatchChangesConnectionResolver = &batchChangesConnectionResolver{}

type batchChangesConnectionResolver struct {
	store           *store.Store
	opts            store.ListBatchChangesOpts
	gitserverClient gitserver.Client
	logger          log.Logger

	// cache results because they are used by multiple fields
	once         sync.Once
	batchChanges []*btypes.BatchChange
	next         int64
	err          error
}

func (r *batchChangesConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.BatchChangeResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.BatchChangeResolver, 0, len(nodes))
	for _, c := range nodes {
		resolvers = append(resolvers, &batchChangeResolver{store: r.store, gitserverClient: r.gitserverClient, batchChange: c, logger: r.logger})
	}
	return resolvers, nil
}

func (r *batchChangesConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := store.CountBatchChangesOpts{
		ChangesetID:                   r.opts.ChangesetID,
		States:                        r.opts.States,
		OnlyAdministeredByUserID:      r.opts.OnlyAdministeredByUserID,
		NamespaceUserID:               r.opts.NamespaceUserID,
		NamespaceOrgID:                r.opts.NamespaceOrgID,
		RepoID:                        r.opts.RepoID,
		ExcludeDraftsNotOwnedByUserID: r.opts.ExcludeDraftsNotOwnedByUserID,
	}
	count, err := r.store.CountBatchChanges(ctx, opts)
	return int32(count), err
}

func (r *batchChangesConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *batchChangesConnectionResolver) compute(ctx context.Context) ([]*btypes.BatchChange, int64, error) {
	r.once.Do(func() {
		r.batchChanges, r.next, r.err = r.store.ListBatchChanges(ctx, r.opts)
	})
	return r.batchChanges, r.next, r.err
}
