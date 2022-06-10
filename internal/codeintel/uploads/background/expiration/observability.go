package expiration

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type metrics struct {
	// Data retention metrics
	numRepositoriesScanned prometheus.Counter
	numUploadsExpired      prometheus.Counter
	numUploadsScanned      prometheus.Counter
	numCommitsScanned      prometheus.Counter
}

var NewMetrics = newMetrics

func newMetrics(observationContext *observation.Context) *metrics {
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

	return &metrics{
		numRepositoriesScanned: numRepositoriesScanned,
		numUploadsScanned:      numUploadsScanned,
		numCommitsScanned:      numCommitsScanned,
		numUploadsExpired:      numUploadsExpired,
	}
}
