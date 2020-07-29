package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var testMetricWarning = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "observability_sample_metric_warning",
	Help: "Value is 1 if warning test alert should be firing, 0 otherwise.",
}, nil)

var testMetricCritical = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "observability_sample_metric_critical",
	Help: "Value is 1 if critical test alert should be firing, 0 otherwise.",
}, nil)

func init() {
	prometheus.MustRegister(testMetricWarning)
	prometheus.MustRegister(testMetricCritical)
}

func (r *schemaResolver) SetObservabilityTestAlertState(ctx context.Context, args *struct {
	Level  string
	Firing bool
}) (*EmptyResponse, error) {
	var metric *prometheus.GaugeVec
	switch args.Level {
	case "warning":
		metric = testMetricWarning
	case "critical":
		metric = testMetricCritical
	default:
		return nil, fmt.Errorf("invalid alert level %q", args.Level)
	}
	if args.Firing {
		metric.With(nil).Set(1)
	} else {
		metric.With(nil).Set(0)
	}
	return &EmptyResponse{}, nil
}
