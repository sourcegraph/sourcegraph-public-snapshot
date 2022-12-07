package background

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type loaderMetrics struct {
	numCSVFilesProcessed prometheus.Counter
	numCSVBytesRead      prometheus.Counter
}

func newLoaderMetrics(observationCtx *observation.Context) *loaderMetrics {
	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationCtx.Registerer.MustRegister(counter)
		return counter
	}

	numCSVFilesProcessed := counter(
		"src_codeintel_ranking_csv_files_processed_total",
		"The number of input CSV records read from GCS.",
	)
	numCSVBytesRead := counter(
		"src_codeintel_ranking_csv_files_bytes_read_total",
		"The number of bytes read from GCS.",
	)

	return &loaderMetrics{
		numCSVFilesProcessed: numCSVFilesProcessed,
		numCSVBytesRead:      numCSVBytesRead,
	}
}
