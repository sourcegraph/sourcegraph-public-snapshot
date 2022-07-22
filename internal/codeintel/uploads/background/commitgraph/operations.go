package commitgraph

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewOperations(uploadSvc UploadService, observationContext *observation.Context) {
	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_total",
		Help: "Total number of repositories with stale commit graphs.",
	}, func() float64 {
		dirtyRepositories, err := uploadSvc.GetDirtyRepositories(context.Background())
		if err != nil {
			observationContext.Logger.Error("Failed to determine number of dirty repositories", log.Error(err))
		}

		return float64(len(dirtyRepositories))
	}))

	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_queued_duration_seconds_total",
		Help: "The maximum amount of time a repository has had a stale commit graph.",
	}, func() float64 {
		age, err := uploadSvc.GetRepositoriesMaxStaleAge(context.Background())
		if err != nil {
			observationContext.Logger.Error("Failed to determine stale commit graph age", log.Error(err))
			return 0
		}

		return float64(age) / float64(time.Second)
	}))
}
