package metrics

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(observationContext *observation.Context, queueOptions map[string]handler.QueueOptions) error {
	// Emit metrics to control alerts
	initPrometheusMetrics(observationContext, queueOptions)

	// Emit metrics to control executor auto-scaling
	return initGCPMetrics(gcpConfig, queueOptions)
}
