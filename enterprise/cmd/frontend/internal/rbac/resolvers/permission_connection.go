package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var _ graphqlbackend.PermissionConnectionResolver = &permissionConnectionResolver{}

type permissionConnectionResolver struct {
	db   database.DB
	opts database.PermissionListOpts

	// cache results because they are used by multiple fields
	once        sync.Once
	permissions []*types.Permission
	totalCount  int
	err         error
}

func (r *permissionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, count, err := r.compute(ctx)
	return int32(count), err
}

func (r *permissionConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	permissions, totalCount, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	// We would have had all results when no limit set
	if r.opts.LimitOffset == nil {
		return graphqlutil.HasNextPage(false), nil
	}

	after := r.opts.LimitOffset.Offset + len(permissions)

	// We got less results than limit, means we've had all results
	if after < r.opts.Limit {
		return graphqlutil.HasNextPage(false), nil
	}

	if totalCount > after {
		return graphqlutil.NextPageCursor(strconv.Itoa(after)), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *permissionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.PermissionResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.PermissionResolver, 0, len(nodes))
	for _, n := range nodes {
		resolvers = append(resolvers, &permissionResolver{
			permission: n,
		})
	}

	return resolvers, nil
}

func (r *permissionConnectionResolver) compute(ctx context.Context) ([]*types.Permission, int, error) {
	r.once.Do(func() {
		r.permissions, r.err = r.db.Permissions().List(ctx, r.opts)
		if r.err != nil {
			return
		}
		r.totalCount, r.err = r.db.Permissions().Count(ctx, r.opts)
		if r.err != nil {
			return
		}
	})

	return r.permissions, r.totalCount, r.err
}
