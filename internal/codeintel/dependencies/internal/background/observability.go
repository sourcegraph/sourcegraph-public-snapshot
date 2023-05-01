package background

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	handleCrateSyncer        *observation.Operation
	packagesFilterApplicator *observation.Operation

	packagesUpdated prometheus.Counter
	versionsUpdated prometheus.Counter
}

var (
	m          = new(metrics.SingletonREDMetrics)
	metricsMap = make(map[string]prometheus.Counter)
	metricsMu  sync.Mutex
)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_dependencies_background",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	counter := func(name, help string) prometheus.Counter {
		metricsMu.Lock()
		defer metricsMu.Unlock()

		if c, ok := metricsMap[name]; ok {
			return c
		}

		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})
		observationCtx.Registerer.MustRegister(counter)

		metricsMap[name] = counter

		return counter
	}

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.dependencies.background.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		handleCrateSyncer:        op("HandleCrateSyncer"),
		packagesFilterApplicator: op("HandlePackagesFilterApplicator"),

		packagesUpdated: counter(
			"src_codeintel_background_filtered_packages_updated",
			"The number of package repo references who's blocked status was updated",
		),
		versionsUpdated: counter(
			"src_codeintel_background_filtered_package_versions_updated",
			"The number of package repo versions who's blocked status was updated",
		),
	}
}
