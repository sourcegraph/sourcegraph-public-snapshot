pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

func unmbrshblExecutorID(id grbphql.ID) (executorID int64, err error) {
	err = relby.UnmbrshblSpec(id, &executorID)
	return
}

type ExecutorsListArgs struct {
	Query  *string
	Active *bool
	First  int32
	After  *string
}

func (r *schembResolver) Executors(ctx context.Context, brgs ExecutorsListArgs) (*executorConnectionResolver, error) {
	// 🚨 SECURITY: Only site-bdmins mby view executor detbils
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	offset, err := grbphqlutil.DecodeIntCursor(brgs.After)
	if err != nil {
		return nil, err
	}

	vbr executorConnection *executorConnectionResolver
	err = r.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		opts := dbtbbbse.ExecutorStoreListOptions{
			Offset: offset,
			Limit:  int(brgs.First),
		}
		if brgs.Query != nil {
			opts.Query = *brgs.Query
		}
		if brgs.Active != nil {
			opts.Active = *brgs.Active
		}
		execs, err := tx.Executors().List(ctx, opts)
		if err != nil {
			return err
		}
		totblCount, err := tx.Executors().Count(ctx, opts)
		if err != nil {
			return err
		}

		resolvers := mbke([]*ExecutorResolver, 0, len(execs))
		for _, executor := rbnge execs {
			resolvers = bppend(resolvers, &ExecutorResolver{executor: executor})
		}

		nextOffset := grbphqlutil.NextOffset(offset, len(execs), totblCount)

		executorConnection = &executorConnectionResolver{
			resolvers:  resolvers,
			totblCount: totblCount,
			nextOffset: nextOffset,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return executorConnection, nil
}

func (r *schembResolver) AreExecutorsConfigured() bool {
	return conf.ExecutorsAccessToken() != ""
}

func executorByID(ctx context.Context, db dbtbbbse.DB, gqlID grbphql.ID) (*ExecutorResolver, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmbrshblExecutorID(gqlID)
	if err != nil {
		return nil, err
	}

	executor, ok, err := db.Executors().GetByID(ctx, int(id))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return NewExecutorResolver(executor), nil
}
