package attribution

import (
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"go.opentelemetry.io/otel/attribute"
)

type operations struct {
	snippetAttribution       *observation.Operation
	snippetAttributionLocal  *observation.Operation
	snippetAttributionDotCom *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"guardrails",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("Guardrails.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	return &operations{
		snippetAttribution:       op("SnippetAttribution"),
		snippetAttributionLocal:  op("SnippetAttributionLocal"),
		snippetAttributionDotCom: op("SnippetAttributionDotCom"),
	}
}

// endObservationWithResult is a helper which will automatically include the
// results logging attribute if it is non-nil.
func endObservationWithResult(traceLogger observation.TraceLogger, endObservation observation.FinishFunc, result **SnippetAttributions) func() {
	// While this feature is experimental we also debug log successful
	// requests. We need to independently capture duration.
	start := time.Now()

	return func() {
		var args observation.Args
		final := *result
		if final != nil {
			args.Attrs = []attribute.KeyValue{
				attribute.Int("len", len(final.RepositoryNames)),
				attribute.Int("total_count", final.TotalCount),
				attribute.Bool("limit_hit", final.LimitHit),
			}

			// Temporary logging code, so duplication is fine with above.
			traceLogger.Debug("successful snippet attribution search",
				log.Int("len", len(final.RepositoryNames)),
				log.Int("total_count", final.TotalCount),
				log.Bool("limit_hit", final.LimitHit),
				log.Duration("duration", time.Since(start)),
			)
		}
		endObservation(1, args)
	}
}
