package local

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

var Orgs sourcegraph.OrgsServer = &orgs{}

type orgs struct{}

var _ sourcegraph.OrgsServer = (*orgs)(nil)

func (s *orgs) Get(ctx context.Context, org *sourcegraph.OrgSpec) (*sourcegraph.Org, error) {
	// TODO: remove when Orgs is implemented
	return nil, grpc.Errorf(codes.Unimplemented, "Orgs is not implemented")

	orgsStore := store.OrgsFromContext(ctx)
	return orgsStore.Get(ctx, *org)
}

func (s *orgs) List(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
	// TODO: remove when Orgs is implemented
	return nil, grpc.Errorf(codes.Unimplemented, "Orgs is not implemented")

	orgsStore := store.OrgsFromContext(ctx)

	orgs, err := orgsStore.List(ctx, op.Member, &op.ListOptions)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.OrgList{Orgs: orgs}, nil
}

func (s *orgs) ListMembers(ctx context.Context, op *sourcegraph.OrgsListMembersOp) (*sourcegraph.UserList, error) {
	// TODO: remove when Orgs is implemented
	return nil, grpc.Errorf(codes.Unimplemented, "Orgs is not implemented")

	orgsStore := store.OrgsFromContext(ctx)

	members, err := orgsStore.ListMembers(ctx, op.Org, op.Opt)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.UserList{Users: members}, nil
}
