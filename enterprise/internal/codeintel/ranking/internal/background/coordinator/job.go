package coordinator

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/cronexpr"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
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
		name,
		"Coordinates the state of the file reference count map and reduce jobs.",
		config.Interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
				return nil
			}

			previous := time.Now() // TODO - from db
			if previous.IsZero() {

				// TODO - bump
				fmt.Printf("NOTHING YET\n")
			} else {
				expr, err := cronexpr.Parse(conf.CodeIntelRankingDocumentReferenceCountsCronExpression())
				if err != nil {
					return err
				}

				if !time.Now().Before(expr.Next(previous)) {
					// TODO - bump
					fmt.Printf("BUMP FROM CRONNIE\n")
				}
			}

			derivativeGraphKeyPrefix, err := store.DerivativeGraphKey(ctx, s)
			if err != nil {
				return err
			}

			return s.Coordinate(ctx, rankingshared.DerivativeGraphKeyFromTime(derivativeGraphKeyPrefix, time.Now()))
		}),
	)
}
