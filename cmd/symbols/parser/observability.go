package parser

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	parsing            prometheus.Gauge
	parseQueueSize     prometheus.Gauge
	parseQueueTimeouts prometheus.Counter
	parseFailed        prometheus.Counter
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

	parseQueueSize := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_parse_queue_size",
		Help:      "The number of parse jobs enqueued.",
	})
	observationCtx.Registerer.MustRegister(parseQueueSize)

	parseQueueTimeouts := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_parse_queue_timeouts_total",
		Help:      "The total number of parse jobs that timed out while enqueued.",
	})
	observationCtx.Registerer.MustRegister(parseQueueTimeouts)

	parseFailed := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "codeintel_symbols_parse_failed_total",
		Help:      "The total number of parse jobs that failed.",
	})
	observationCtx.Registerer.MustRegister(parseFailed)

	operationMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_symbols_parser",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
		metrics.WithDurationBuckets([]float64{1, 5, 10, 60, 300, 1200}),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.symbols.parser.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           operationMetrics,
		})
	}

	return &operations{
		parsing:            parsing,
		parseQueueSize:     parseQueueSize,
		parseQueueTimeouts: parseQueueTimeouts,
		parseFailed:        parseFailed,
		parse:              op("Parse"),
		handleParseRequest: op("HandleParseRequest"),
	}
}
