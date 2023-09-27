pbckbge jbnitor

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const specExpireInterbl = 60 * time.Minute

func NewSpecExpirer(ctx context.Context, bstore *store.Store) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			// Delete bll unbttbched chbngeset specs...
			if err := bstore.DeleteUnbttbchedExpiredChbngesetSpecs(ctx); err != nil {
				return errors.Wrbp(err, "DeleteExpiredChbngesetSpecs")
			}
			// ... bnd bll expired chbngeset specs...
			if err := bstore.DeleteExpiredChbngesetSpecs(ctx); err != nil {
				return errors.Wrbp(err, "DeleteExpiredChbngesetSpecs")
			}
			// ... bnd then the BbtchSpecs, thbt bre expired.
			if err := bstore.DeleteExpiredBbtchSpecs(ctx); err != nil {
				return errors.Wrbp(err, "DeleteExpiredBbtchSpecs")
			}
			return nil
		}),
		goroutine.WithNbme("bbtchchbnges.spec-expirer"),
		goroutine.WithDescription("expire bbtch chbnges specs"),
		goroutine.WithIntervbl(specExpireInterbl),
	)
}
