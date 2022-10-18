package background

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ExpirationMetrics struct {
	// Data retention metrics
	NumRepositoriesScanned prometheus.Counter
	NumUploadsExpired      prometheus.Counter
	NumUploadsScanned      prometheus.Counter
	NumCommitsScanned      prometheus.Counter
}

func NewExpirationMetrics(observationContext *observation.Context) *ExpirationMetrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationContext.Registerer.MustRegister(counter)
		return counter
	}

	numRepositoriesScanned := counter(
		"src_codeintel_background_repositories_scanned_total",
		"The number of repositories scanned for data retention.",
	)
	numUploadsScanned := counter(
		"src_codeintel_background_upload_records_scanned_total",
		"The number of codeintel upload records scanned for data retention.",
	)
	numCommitsScanned := counter(
		"src_codeintel_background_commits_scanned_total",
		"The number of commits reachable from a codeintel upload record scanned for data retention.",
	)
	numUploadsExpired := counter(
		"src_codeintel_background_upload_records_expired_total",
		"The number of codeintel upload records marked as expired.",
	)

	return &ExpirationMetrics{
		NumRepositoriesScanned: numRepositoriesScanned,
		NumUploadsScanned:      numUploadsScanned,
		NumCommitsScanned:      numCommitsScanned,
		NumUploadsExpired:      numUploadsExpired,
	}
}

func (b backgroundJob) SetMetricReporters(observationContext *observation.Context) {
	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_total",
		Help: "Total number of repositories with stale commit graphs.",
	}, func() float64 {
		dirtyRepositories, err := b.uploadSvc.GetDirtyRepositories(context.Background())
		if err != nil {
			observationContext.Logger.Error("Failed to determine number of dirty repositories", log.Error(err))
		}

		return float64(len(dirtyRepositories))
	}))

	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_queued_duration_seconds_total",
		Help: "The maximum amount of time a repository has had a stale commit graph.",
	}, func() float64 {
		age, err := b.uploadSvc.GetRepositoriesMaxStaleAge(context.Background())
		if err != nil {
			observationContext.Logger.Error("Failed to determine stale commit graph age", log.Error(err))
			return 0
		}

		return float64(age) / float64(time.Second)
	}))
}
