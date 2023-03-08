package uploads

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	inferClosestUploads *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_uploads",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		inferClosestUploads: op("InferClosestUploads"),
	}
}

func MetricReporters(observationCtx *observation.Context, uploadSvc UploadService) {
	observationCtx.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_total",
		Help: "Total number of repositories with stale commit graphs.",
	}, func() float64 {
		dirtyRepositories, err := uploadSvc.GetDirtyRepositories(context.Background())
		if err != nil {
			observationCtx.Logger.Error("Failed to determine number of dirty repositories", log.Error(err))
		}

		return float64(len(dirtyRepositories))
	}))

	observationCtx.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_queued_duration_seconds_total",
		Help: "The maximum amount of time a repository has had a stale commit graph.",
	}, func() float64 {
		age, err := uploadSvc.GetRepositoriesMaxStaleAge(context.Background())
		if err != nil {
			observationCtx.Logger.Error("Failed to determine stale commit graph age", log.Error(err))
			return 0
		}

		return float64(age) / float64(time.Second)
	}))
}
