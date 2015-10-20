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

	// TODO: Do I need to error check that key is not nil?
	//       Why is it a pointer anyway, can't we get grpc to generate it as a value?
	err := store.AddKey(ctx, int32(actor.UID), *key)
	if err != nil {
		return nil, err
	}

	noCache(ctx) // TODO: What should the cache be?
	return &pbtypes.Void{}, nil
}

func (s *userKeys) LookupUser(ctx context.Context, key *sourcegraph.SSHPublicKey) (*sourcegraph.UserSpec, error) {
	// TODO: Consider not requiring write access for lookup?
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "UserKeys.LookupUser"); err != nil {
		return nil, err
	}

	store := store.UserKeysFromContextOrNil(ctx)
	if store == nil {
		return nil, &sourcegraph.NotImplementedError{What: "UserKeys"}
	}

	userSpec, err := store.LookupUser(ctx, *key)
	if err != nil {
		return nil, err
	}

	noCache(ctx) // TODO: What should the cache be?
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
