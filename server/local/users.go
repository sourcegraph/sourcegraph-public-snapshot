package local

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

var Users sourcegraph.UsersServer = &users{}

type users struct{}

var _ sourcegraph.UsersServer = (*users)(nil)

func (s *users) Get(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.User, error) {
	store := store.UsersFromContextOrNil(ctx)
	if store == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "Users")
	}

	shortCache(ctx)
	return store.Get(ctx, *user)
}

func (s *users) GetWithEmail(ctx context.Context, emailAddr *sourcegraph.EmailAddr) (*sourcegraph.User, error) {
	store := store.UsersFromContextOrNil(ctx)
	if store == nil {
		return nil, grpc.Errorf(codes.Unimplemented, "Users")
	}

	shortCache(ctx)
	return store.GetWithEmail(ctx, *emailAddr)
}

func (s *users) ListEmails(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.EmailAddrList, error) {
	store := store.UsersFromContextOrNil(ctx)
	if store == nil {
		log.Printf("Warning: users not implemented, returning empty list")
		return &sourcegraph.EmailAddrList{}, nil
	}

	emails, err := store.ListEmails(ctx, *user)
	if err != nil {
		return nil, err
	}
	shortCache(ctx)
	return &sourcegraph.EmailAddrList{EmailAddrs: emails}, nil
}

func (s *users) List(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
	store := store.UsersFromContextOrNil(ctx)
	if store == nil {
		log.Printf("Warning: users not implemented, returning empty list")
		return &sourcegraph.UserList{}, nil
	}

	users, err := store.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	shortCache(ctx)
	return &sourcegraph.UserList{Users: users}, nil
}

func (s *users) Count(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.UserCount, error) {
	noCache(ctx)

	store := store.UsersFromContextOrNil(ctx)
	if store == nil {
		log.Printf("Warning: users not implemented, returning zero")
		return &sourcegraph.UserCount{}, nil
	}

	count, err := store.Count(elevatedActor(ctx))
	if err != nil {
		return nil, err
	}

	if count > 0 {
		// If the request is not authed with admin privileges, don't reveal the actual
		// number of users.
		if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Users.Count"); err != nil {
			count = 1729 // https://en.wikipedia.org/wiki/Taxicab_number
		}
	}
	return &sourcegraph.UserCount{Count: count}, nil
}
