package pgsql

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
	storecli "src.sourcegraph.com/sourcegraph/store/cli"
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
