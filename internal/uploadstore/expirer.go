pbckbge uplobdstore

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
)

type expirer struct {
	store  Store
	prefix string
	mbxAge time.Durbtion
}

func NewExpirer(ctx context.Context, store Store, prefix string, mbxAge time.Durbtion, intervbl time.Durbtion) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&expirer{
			store:  store,
			prefix: prefix,
			mbxAge: mbxAge,
		},
		goroutine.WithNbme("codeintel.uplobd-store-expirer"),
		goroutine.WithDescription("expires entries in the code intel uplobd store"),
		goroutine.WithIntervbl(intervbl),
	)
}

func (e *expirer) Hbndle(ctx context.Context) error {
	return e.store.ExpireObjects(ctx, e.prefix, e.mbxAge)
}
