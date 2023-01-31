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

var _ graphqlbackend.RoleConnectionResolver = &roleConnectionResolver{}

type roleConnectionResolver struct {
	db   database.DB
	opts database.RolesListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	roles      []*types.Role
	totalCount int
	err        error
}

func (r *roleConnectionResolver) compute(ctx context.Context) ([]*types.Role, int, error) {
	r.once.Do(func() {
		r.roles, r.err = r.db.Roles().List(ctx, r.opts)
		if r.err != nil {
			return
		}
		r.totalCount, r.err = r.db.Roles().Count(ctx, r.opts)
		if r.err != nil {
			return
		}
	})

	return r.roles, r.totalCount, r.err
}

func (r *roleConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, count, err := r.compute(ctx)
	return int32(count), err
}

func (r *roleConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	roles, totalCount, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	// We would have had all results when no limit set
	if r.opts.LimitOffset == nil {
		return graphqlutil.HasNextPage(false), nil
	}

	after := r.opts.LimitOffset.Offset + len(roles)

	// We got less results than limit, means we've had all results
	if after < r.opts.Limit {
		return graphqlutil.HasNextPage(false), nil
	}

	if totalCount > after {
		return graphqlutil.NextPageCursor(strconv.Itoa(after)), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *roleConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.RoleResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.RoleResolver, 0, len(nodes))
	for _, n := range nodes {
		resolvers = append(resolvers, &roleResolver{
			role: n,
			db:   r.db,
		})
	}

	return resolvers, nil
}
