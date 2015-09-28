package pgsql

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/server/serverctx"
	storecli "sourcegraph.com/sourcegraph/sourcegraph/store/cli"
)

func init() {
	// Make the DB handle available in the server's context.
	serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
		if storecli.ActiveFlags.Store == "pgsql" {
			return NewContext(ctx, DB()), nil
		}
		return ctx, nil
	})
}
