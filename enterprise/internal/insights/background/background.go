package background

import (
	"context"
	"database/sql"
	"log"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func StartBackgroundJobs(ctx context.Context, db *sql.DB) {
	resolver, err := insights.InitResolver(ctx, db)
	if err != nil {
		// e.g. migration failed, DB unavailable, etc. code insights is non-functional so we do not
		// want to continue.
		log.Fatal("failed to initialize code insights (set DISABLE_CODE_INSIGHTS=true if needed)", err)
	}

	// Create metrics for recording information about background jobs.
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	metrics := newMetrics(observationContext)

	// Start background goroutines for all of our workers.
	routines := []goroutine.BackgroundRoutine{
		newInsightEnqueuer(ctx, resolver.Store),
		newQueryRunner(ctx, resolver.Store, resolver, metrics), // TODO(slimsag): should not store in TimescaleDB
		newQueryRunnerResetter(ctx, resolver.Store, metrics),   // TODO(slimsag): should not store in TimescaleDB
	}
	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
}
