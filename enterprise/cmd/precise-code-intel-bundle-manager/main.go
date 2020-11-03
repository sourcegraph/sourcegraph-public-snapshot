package main

import (
	"context"
	"log"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/server"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

var bundleDir = env.Get("PRECISE_CODE_INTEL_BUNDLE_DIR", "/lsif-storage", "Root dir containing uploads and converted bundles.")

func main() {
	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	tracer.Init()
	trace.Init(true)

	if bundleDir == "" {
		log.Fatalf("invalid value %q for %s: no value supplied", bundleDir, "PRECISE_CODE_INTEL_BUNDLE_DIR")
	}

	sqliteutil.MustRegisterSqlite3WithPcre()

	if err := paths.PrepDirectories(bundleDir); err != nil {
		log.Fatalf("failed to prepare directories: %s", err)
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	debugServer, err := debugserver.NewServerRoutine()
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}
	go debugServer.Start()

	server, err := server.New(bundleDir, observationContext)
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}

	goroutine.MonitorBackgroundRoutines(context.Background(), server)
}
