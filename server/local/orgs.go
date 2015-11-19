package local

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

var Orgs sourcegraph.OrgsServer = &orgs{}

type orgs struct{}

var _ sourcegraph.OrgsServer = (*orgs)(nil)

func (s *orgs) Get(ctx context.Context, org *sourcegraph.OrgSpec) (*sourcegraph.Org, error) {
	orgsStore := store.OrgsFromContextOrNil(ctx)
	if orgsStore == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "no Orgs")
	}
	return orgsStore.Get(ctx, *org)
}

func (s *orgs) List(ctx context.Context, op *sourcegraph.OrgsListOp) (*sourcegraph.OrgList, error) {
	orgsStore := store.OrgsFromContextOrNil(ctx)
	if orgsStore == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "no Orgs")
	}

	orgs, err := orgsStore.List(ctx, op.Member, &op.ListOptions)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.OrgList{Orgs: orgs}, nil
}

func (s *orgs) ListMembers(ctx context.Context, op *sourcegraph.OrgsListMembersOp) (*sourcegraph.UserList, error) {
	orgsStore := store.OrgsFromContextOrNil(ctx)
	if orgsStore == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "no Orgs")
	}

	members, err := orgsStore.ListMembers(ctx, op.Org, op.Opt)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.UserList{Users: members}, nil
}
