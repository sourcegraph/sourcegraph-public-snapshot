package local

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

var UserKeys sourcegraph.UserKeysServer = &userKeys{}

type userKeys struct{}

var _ sourcegraph.UserKeysServer = (*userKeys)(nil)

func (s *userKeys) AddKey(ctx context.Context, key *sourcegraph.SSHPublicKey) (*pbtypes.Void, error) {
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	store := store.UserKeysFromContextOrNil(ctx)
	if store == nil {
		return nil, &sourcegraph.NotImplementedError{What: "UserKeys"}
	}

	err := store.AddKey(ctx, int32(actor.UID), *key)
	if err != nil {
		return nil, err
	}

	noCache(ctx)
	return &pbtypes.Void{}, nil
}

// LookupUser looks up user by key. The returned UserSpec will only have UID field set.
func (s *userKeys) LookupUser(ctx context.Context, key *sourcegraph.SSHPublicKey) (*sourcegraph.UserSpec, error) {
	store := store.UserKeysFromContextOrNil(ctx)
	if store == nil {
		return nil, &sourcegraph.NotImplementedError{What: "UserKeys"}
	}

	userSpec, err := store.LookupUser(ctx, *key)
	if err != nil {
		return nil, err
	}

	mediumCache(ctx) // Note, after a user deletes or changes their key, it'll not have effect for duration of cache.
	return userSpec, nil
}

func (s *userKeys) DeleteKey(ctx context.Context, _ *pbtypes.Void) (*pbtypes.Void, error) {
	actor := authpkg.ActorFromContext(ctx)

	if !actor.IsAuthenticated() {
		return nil, grpc.Errorf(codes.PermissionDenied, "no authenticated user in context")
	}

	store := store.UserKeysFromContextOrNil(ctx)
	if store == nil {
		return nil, &sourcegraph.NotImplementedError{What: "UserKeys"}
	}

	err := store.DeleteKey(ctx, int32(actor.UID))
	if err != nil {
		return nil, err
	}

	noCache(ctx)
	return &pbtypes.Void{}, nil
}
