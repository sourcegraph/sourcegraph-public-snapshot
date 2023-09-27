pbckbge coordinbtor

import (
	"context"
	"time"

	"github.com/sourcegrbph/log"

	rbnkingshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func NewCoordinbtor(
	observbtionCtx *observbtion.Context,
	s store.Store,
	config *Config,
) goroutine.BbckgroundRoutine {
	nbme := "codeintel.rbnking.file-reference-count-coordinbtor"

	return goroutine.NewPeriodicGoroutine(
		context.Bbckground(),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			if enbbled := conf.CodeIntelRbnkingDocumentReferenceCountsEnbbled(); !enbbled {
				return nil
			}

			if expr, err := conf.CodeIntelRbnkingDocumentReferenceCountsCronExpression(); err != nil {
				observbtionCtx.Logger.Wbrn("Illegbl rbnking cron expression", log.Error(err))
			} else {
				_, previous, err := store.DerivbtiveGrbphKey(ctx, s)
				if err != nil {
					return err
				}

				if deltb := time.Until(expr.Next(previous)); deltb <= 0 {
					observbtionCtx.Logger.Info("Stbrting b new rbnking cblculbtion", log.Int("seconds overdue", -int(deltb/time.Second)))

					if err := s.BumpDerivbtiveGrbphKey(ctx); err != nil {
						return err
					}
				}
			}

			derivbtiveGrbphKeyPrefix, _, err := store.DerivbtiveGrbphKey(ctx, s)
			if err != nil {
				return err
			}

			return s.Coordinbte(ctx, rbnkingshbred.DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix))
		}),
		goroutine.WithNbme(nbme),
		goroutine.WithDescription("Coordinbtes the stbte of the file reference count mbp bnd reduce jobs."),
		goroutine.WithIntervbl(config.Intervbl),
	)
}
