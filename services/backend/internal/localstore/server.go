package localstore

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/serverctx"
)

func init() {
	// Make the DB handle available in the server's context.
	serverctx.Funcs = append(serverctx.Funcs, func(ctx context.Context) (context.Context, error) {
		appDBH, graphDBH, err := globalDBs()
		if err != nil {
			return nil, err
		}
		dbCtx := WithAppDBH(ctx, appDBH)
		dbCtx = WithGraphDBH(dbCtx, graphDBH)
		return dbCtx, nil
	})
}
