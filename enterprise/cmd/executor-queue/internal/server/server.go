package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// metricsScrapeInterval denotes the delay between two runs of collecting the aggregated
// metrics data from all registered handlers.
const metricsScrapeInterval = 10 * time.Second

// NewServer returns an HTTP job queue server.
func NewServer(options Options, queueOptions map[string]QueueOptions, observationContext *observation.Context) goroutine.BackgroundRoutine {
	queueMetrics := newQueueMetrics(observationContext)

	addr := fmt.Sprintf(":%d", options.Port)
	scraper, router := setupRoutes(options, queueOptions)
	httpHandler := ot.Middleware(httpserver.NewHandler(router))
	server := httpserver.NewFromAddr(addr, &http.Server{Handler: httpHandler})

	logScraper := goroutine.NewPeriodicGoroutine(
		context.Background(),
		metricsScrapeInterval,
		goroutine.NewHandlerWithErrorMessage("scrape metrics from queue handlers", func(ctx context.Context) error {
			for name, metrics := range scraper() {
				queueMetrics.NumJobs.WithLabelValues(name).Set(float64(metrics.NumJobs))
				queueMetrics.NumExecutors.WithLabelValues(name).Set(float64(metrics.NumExecutors))
			}
			return nil
		}),
	)

	routines := goroutine.CombinedRoutine{server, logScraper}

	return routines
}
