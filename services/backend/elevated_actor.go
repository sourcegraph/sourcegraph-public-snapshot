package backend

import (
	"context"

	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

// elevatedActor returns an actor with admin access to the stores.
//
// CAUTION: use this function only in cases where it is required
// to complete an operation with elevated access, for example when
// creating an account when a user signs up. DO NOT USE this actor
// to complete requests that will return store data in the response.
func elevatedActor(ctx context.Context) context.Context {
	return authpkg.WithActor(ctx, authpkg.Actor{Scope: map[string]bool{"internal:elevated": true}})
}
