package uploads

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type expirationMetrics struct {
	// Data retention metrics
	numRepositoriesScanned prometheus.Counter
	numUploadsExpired      prometheus.Counter
	numUploadsScanned      prometheus.Counter
	numCommitsScanned      prometheus.Counter
}

func newExpirationMetrics(observationContext *observation.Context) *expirationMetrics {
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

	return &expirationMetrics{
		numRepositoriesScanned: numRepositoriesScanned,
		numUploadsScanned:      numUploadsScanned,
		numCommitsScanned:      numCommitsScanned,
		numUploadsExpired:      numUploadsExpired,
	}
}

func (s *Service) MetricReporters(observationContext *observation.Context) {
	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_total",
		Help: "Total number of repositories with stale commit graphs.",
	}, func() float64 {
		dirtyRepositories, err := s.store.GetDirtyRepositories(context.Background())
		if err != nil {
			observationContext.Logger.Error("Failed to determine number of dirty repositories", log.Error(err))
		}

		return float64(len(dirtyRepositories))
	}))

	observationContext.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_queued_duration_seconds_total",
		Help: "The maximum amount of time a repository has had a stale commit graph.",
	}, func() float64 {
		age, err := s.store.GetRepositoriesMaxStaleAge(context.Background())
		if err != nil {
			observationContext.Logger.Error("Failed to determine stale commit graph age", log.Error(err))
			return 0
		}

		return float64(age) / float64(time.Second)
	}))
}
