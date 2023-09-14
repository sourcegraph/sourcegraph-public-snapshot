package coordinator

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewCoordinator(
	observationCtx *observation.Context,
	s store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.file-reference-count-coordinator"

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
				return nil
			}

			if expr, err := conf.CodeIntelRankingDocumentReferenceCountsCronExpression(); err != nil {
				observationCtx.Logger.Warn("Illegal ranking cron expression", log.Error(err))
			} else {
				_, previous, err := store.DerivativeGraphKey(ctx, s)
				if err != nil {
					return err
				}

				if delta := time.Until(expr.Next(previous)); delta <= 0 {
					observationCtx.Logger.Info("Starting a new ranking calculation", log.Int("seconds overdue", -int(delta/time.Second)))

					if err := s.BumpDerivativeGraphKey(ctx); err != nil {
						return err
					}
				}
			}

			derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
			if err != nil {
				return err
			}

			return s.Coordinate(ctx, rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix))
		}),
		goroutine.WithName(name),
		goroutine.WithDescription("Coordinates the state of the file reference count map and reduce jobs."),
		goroutine.WithInterval(config.Interval),
	)
}
