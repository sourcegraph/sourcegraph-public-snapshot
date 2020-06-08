package indexer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
)

type IndexerMetrics struct {
	Processor *metrics.OperationMetrics
}

func NewIndexerMetrics(r prometheus.Registerer) IndexerMetrics {
	processor := metrics.NewOperationMetrics(r, "index_queue_processor")

	return IndexerMetrics{
		Processor: processor,
	}
}
