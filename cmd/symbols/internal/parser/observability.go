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

func newOperations(observationContext *observation.Context) *operations {
	parsing := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "symbols_parse_parsing",
		Help: "The number of parse jobs currently running.",
	})
	observationContext.Registerer.MustRegister(parsing)

	parseQueueSize := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "symbols_parse_parse_queue_size",
		Help: "The number of parse jobs enqueued.",
	})
	observationContext.Registerer.MustRegister(parseQueueSize)

	parseQueueTimeouts := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "symbols_parse_parse_queue_timeouts",
		Help: "The total number of parse jobs that timed out while enqueued.",
	})
	observationContext.Registerer.MustRegister(parseQueueTimeouts)

	parseFailed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "symbols_parse_parse_failed",
		Help: "The total number of parse jobs that failed.",
	})
	observationContext.Registerer.MustRegister(parseFailed)

	operationMetrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_symbols_parser",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
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
