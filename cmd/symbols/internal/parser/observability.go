package parser

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type operations struct {
	parsing            prometheus.Gauge
	parseFailed        prometheus.Counter
	parseCanceled      prometheus.Counter
	parse              *observation.Operation
	handleParseRequest *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	parsing := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_parsing",
		Help:      "The number of parse jobs currently running.",
	})
	observationCtx.Registerer.MustRegister(parsing)

	parseFailed := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_parse_failed_total",
		Help:      "The total number of parse jobs that failed.",
	})
	observationCtx.Registerer.MustRegister(parseFailed)

	parseCanceled := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_parse_canceled_total",
		Help:      "The total number of parse jobs that are canceled. Seperate to failed since we don't treat these as failed parses.",
	})
	observationCtx.Registerer.MustRegister(parseCanceled)

	operationMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_symbols_parser",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
		metrics.WithDurationBuckets([]float64{1, 5, 10, 60, 300, 1200}),
	)

	op := func(name string) observation.Op {
		return observation.Op{
			Name:              fmt.Sprintf("codeintel.symbols.parser.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           operationMetrics,
		}
	}

	// HandleParseRequest we run concurrently and when we cancel the request
	// they all logspam the error. But in the case of cancelation this is not
	// useful signal, but rather higher level operations should log the
	// cancelation.
	handleParseRequestOp := op("HandleParseRequest")
	handleParseRequestOp.ErrorFilter = func(err error) observation.ErrorFilterBehaviour {
		if errors.Is(err, context.Canceled) {
			return observation.EmitForAllExceptLogs
		}
		return observation.EmitForDefault
	}

	return &operations{
		parsing:            parsing,
		parseFailed:        parseFailed,
		parseCanceled:      parseCanceled,
		parse:              observationCtx.Operation(op("Parse")),
		handleParseRequest: observationCtx.Operation(handleParseRequestOp),
	}
}
