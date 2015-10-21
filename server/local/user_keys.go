package local

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/server/internal/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

var UserKeys sourcegraph.UserKeysServer = &userKeys{}

type userKeys struct{}

var _ sourcegraph.UserKeysServer = (*userKeys)(nil)

func (s *userKeys) AddKey(ctx context.Context, key *sourcegraph.SSHPublicKey) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "UserKeys.AddKey"); err != nil {
		return nil, err
	}

	store := store.UserKeysFromContextOrNil(ctx)
	if store == nil {
		return nil, &sourcegraph.NotImplementedError{What: "UserKeys"}
	}

	actor := authpkg.ActorFromContext(ctx)

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

	mediumCache(ctx) // TODO: But if a user deletes a key, it'll not have effect for 300 seconds.
	return userSpec, nil
}

func (s *userKeys) DeleteKey(ctx context.Context, _ *pbtypes.Void) (*pbtypes.Void, error) {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "UserKeys.DeleteKey"); err != nil {
		return nil, err
	}

	store := store.UserKeysFromContextOrNil(ctx)
	if store == nil {
		return nil, &sourcegraph.NotImplementedError{What: "UserKeys"}
	}

	actor := authpkg.ActorFromContext(ctx)

	err := store.DeleteKey(ctx, int32(actor.UID))
	if err != nil {
		return nil, err
	}

	noCache(ctx) // TODO: What should the cache be?
	return &pbtypes.Void{}, nil
}
