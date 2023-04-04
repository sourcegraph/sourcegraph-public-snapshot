package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCVEMatcher(store store.Store, metrics *Metrics, interval time.Duration, batchSize int) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"codeintel.sentinel-cve-matcher", "Matches SCIP indexes against known vulnerabilities.",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			numReferencesScanned, numVulnerabilityMatches, err := store.ScanMatches(ctx, batchSize)
			if err != nil {
				return err
			}

			metrics.numReferencesScanned.Add(float64(numReferencesScanned))
			metrics.numVulnerabilityMatches.Add(float64(numVulnerabilityMatches))
			return nil
		}),
	)
}
