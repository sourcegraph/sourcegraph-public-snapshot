package graphqlbackend

import (
	"context"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (r *schemaResolver) AddCodehost(ctx context.Context, args *struct {
	Kind        string
	DisplayName string
	Config      string
}) (*codehostResolver, error) {
	// 🚨 SECURITY: Only site admins may add codehosts.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	codehost := &types.Codehost{
		Kind:        args.Kind,
		DisplayName: args.DisplayName,
		Config:      args.Config,
	}
	err := db.Codehosts.Create(ctx, codehost)
	return &codehostResolver{codehost: codehost}, err
}

func (*schemaResolver) UpdateCodehost(ctx context.Context, args *struct {
	ID          graphql.ID
	DisplayName *string
	Config      *string
}) (*codehostResolver, error) {
	codehostID, err := unmarshalCodehostID(args.ID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site admins are allowed to update the user.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	update := &db.CodehostUpdate{
		DisplayName: args.DisplayName,
		Config:      args.Config,
	}
	if err := db.Codehosts.Update(ctx, codehostID, update); err != nil {
		return nil, err
	}

	codehost, err := db.Codehosts.GetByID(ctx, codehostID)
	if err != nil {
		return nil, err
	}
	return &codehostResolver{codehost: codehost}, nil
}

func (r *schemaResolver) Codehosts(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*codehostConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may read codehost (they have secrets).
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	var opt db.CodehostsListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &codehostConnectionResolver{opt: opt}, nil
}

type codehostConnectionResolver struct {
	opt db.CodehostsListOptions

	// cache results because they are used by multiple fields
	once      sync.Once
	codehosts []*types.Codehost
	err       error
}

func (r *codehostConnectionResolver) compute(ctx context.Context) ([]*types.Codehost, error) {
	r.once.Do(func() {
		r.codehosts, r.err = db.Codehosts.List(ctx, r.opt)
	})
	return r.codehosts, r.err
}

func (r *codehostConnectionResolver) Nodes(ctx context.Context) ([]*codehostResolver, error) {
	codehosts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*codehostResolver, 0, len(codehosts))
	for _, codehost := range codehosts {
		resolvers = append(resolvers, &codehostResolver{codehost: codehost})
	}
	return resolvers, nil
}

func (r *codehostConnectionResolver) TotalCount(ctx context.Context) (countptr int32, err error) {
	count, err := db.Codehosts.Count(ctx, r.opt)
	return int32(count), err
}

func (r *codehostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	codehosts, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(codehosts) >= r.opt.Limit), nil
}
