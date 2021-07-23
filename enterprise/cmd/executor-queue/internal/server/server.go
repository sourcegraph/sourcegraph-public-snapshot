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

// ServerOptions captures the options required for setting up an executor queue
// server.
type ServerOptions struct {
	Port int
}

// NewServer returns an HTTP job queue server.
func NewServer(options ServerOptions, queueOptions map[string]QueueOptions, observationContext *observation.Context) goroutine.BackgroundRoutine {
	addr := fmt.Sprintf(":%d", options.Port)
	router := setupRoutes(queueOptions)
	httpHandler := ot.Middleware(httpserver.NewHandler(router))
	server := httpserver.NewFromAddr(addr, &http.Server{Handler: httpHandler})

	queueMetrics := newQueueMetrics(observationContext)
	logScraper := goroutine.NewPeriodicGoroutine(
		context.Background(),
		metricsScrapeInterval,
		goroutine.NewHandlerWithErrorMessage("scrape metrics from queue handlers", func(ctx context.Context) error {
			for name := range queueOptions {
				// TODO: Reimplement metrics scraping from the database.
				// SELECT
				// 	COUNT(id)
				// FROM
				// 	{table_name}
				// GROUP BY worker_hostname
				// for executorName, meta := range h.executors {
				// 	for _, job := range meta.jobs {
				// 		JobIDs = append(JobIDs, job.recordID)
				// 		ExecutorNames[executorName] = struct{}{}
				// 	}
				// }
				// We don't record executors anymore as we don't hold them in memory,
				// so we cannot tell when whether an executor is not getting a job or it is dead
				// right now. We want to build out a separate store for executors so they're persisted
				// for a while, so we can build some admin UI for viewing their status.
				// Maybe it is fine until then to not have this number and report it again then.
				// Otherwise we need to hold it in the DB now.
				queueMetrics.NumJobs.WithLabelValues(name).Set(float64(0))
				queueMetrics.NumExecutors.WithLabelValues(name).Set(float64(0))
			}
			return nil
		}),
	)

	routines := goroutine.CombinedRoutine{server, logScraper}

	return routines
}
