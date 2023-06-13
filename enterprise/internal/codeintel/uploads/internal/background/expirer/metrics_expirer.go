package expirer

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type ExpirationMetrics struct {
	NumRepositoriesScanned prometheus.Counter
	NumUploadsExpired      prometheus.Counter
	NumUploadsScanned      prometheus.Counter
	NumCommitsScanned      prometheus.Counter
}

var expirationMetrics = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (*ExpirationMetrics, error) {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		r.MustRegister(counter)
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
	}, nil
})

func NewExpirationMetrics(observationCtx *observation.Context) *ExpirationMetrics {
	metrics, _ := expirationMetrics.Init(observationCtx.Registerer)
	return metrics
}
