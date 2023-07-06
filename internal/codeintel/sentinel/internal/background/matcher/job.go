package matcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewCVEMatcher(store store.Store, observationCtx *observation.Context, config *Config) goroutine.BackgroundRoutine {
	metrics := newMetrics(observationCtx)

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(context.Background()),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			numReferencesScanned, numVulnerabilityMatches, err := store.ScanMatches(ctx, config.BatchSize)
			if err != nil {
				return err
			}

			metrics.numReferencesScanned.Add(float64(numReferencesScanned))
			metrics.numVulnerabilityMatches.Add(float64(numVulnerabilityMatches))
			return nil
		}),
		goroutine.WithName("codeintel.sentinel-cve-matcher"),
		goroutine.WithDescription("Matches SCIP indexes against known vulnerabilities."),
		goroutine.WithInterval(config.MatcherInterval),
	)
}
