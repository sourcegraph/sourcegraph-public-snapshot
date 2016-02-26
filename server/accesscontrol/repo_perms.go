package accesscontrol

import (
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/store"
)

// SetWaitlistStatus checks if the PrivateMirrors feature is enabled
// and the actor corresponds to a logged-in user, and sets the
// appropriate waitlist state for the actor.
func SetWaitlistStatus(ctx context.Context, actor *auth.Actor) {
	if !authutil.ActiveFlags.PrivateMirrors || actor == nil || actor.UID == 0 {
		return
	}

	if authutil.ActiveFlags.MirrorsWaitlist != "none" {
		waitlistedUser, err := store.WaitlistFromContext(ctx).GetUser(elevatedActor(ctx), int32(actor.UID))
		if err != nil {
			if _, ok := err.(*store.WaitlistedUserNotFoundError); !ok {
				log15.Debug("Error fetching waitlisted user", "uid", actor.UID, "error", err)
			}
			return
		}

		if waitlistedUser.GrantedAt == nil {
			// User is on the waitlist.
			actor.MirrorsWaitlist = true
			return
		}
	}

	actor.PrivateReposAllowed = true
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
