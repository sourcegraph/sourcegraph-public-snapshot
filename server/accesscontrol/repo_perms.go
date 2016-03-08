package accesscontrol

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/store"
)

// SetWaitlistStatus stores whether the PrivateReposAllowed feature is
// enabled for the actor.
func SetWaitlistStatus(ctx context.Context, actor *auth.Actor) error {
	if actor == nil || actor.UID == 0 {
		return nil
	}

	if authutil.ActiveFlags.MirrorsWaitlist != "none" {
		waitlistedUser, err := store.WaitlistFromContext(ctx).GetUser(elevatedActor(ctx), int32(actor.UID))
		if err != nil {
			if _, ok := err.(*store.WaitlistedUserNotFoundError); !ok {
				return err
			}
			return nil
		}

		if waitlistedUser.GrantedAt == nil {
			// User is on the waitlist. Don't set PrivateReposAllowed.
			return nil
		}
	}

	actor.PrivateReposAllowed = true
	return nil
}

// elevatedActor returns an actor with admin access to the stores.
//
// CAUTION: use this function only in cases where it is required
// to complete an operation with elevated access, for example when
// creating an account when a user signs up. DO NOT USE this actor
// to complete requests that will return store data in the response.
func elevatedActor(ctx context.Context) context.Context {
	return auth.WithActor(ctx, auth.Actor{Scope: map[string]bool{"internal:elevated": true}})
}
