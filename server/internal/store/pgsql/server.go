package pgsql

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
)

func init() {
	// Make the DB handle available in the server's context.
	serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
		dbh, err := globalDB()
		if err != nil {
			return nil, err
		}
		return NewContext(ctx, dbh), nil
	})
}
