package local

import (
	"golang.org/x/net/context"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
)

func elevatedActor(ctx context.Context) context.Context {
	return authpkg.WithActor(ctx, authpkg.Actor{Scope: map[string]bool{"internal:elevated": true}})
}
