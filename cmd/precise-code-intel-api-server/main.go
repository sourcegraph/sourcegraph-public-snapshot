package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/api"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/resetter"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/server"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	var (
		bundleManagerURL = mustGet(rawBundleManagerURL, "PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL")
		resetInterval    = mustParseInterval(rawResetInterval, "PRECISE_CODE_INTEL_RESET_INTERVAL")
	)

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db := db.NewObserved(mustInitializeDatabase(), observationContext)
	bundleManagerClient := bundles.New(bundleManagerURL)
	codeIntelAPI := api.NewObserved(api.New(db, bundleManagerClient, gitserver.DefaultClient), observationContext)
	resetterMetrics := resetter.NewResetterMetrics(prometheus.DefaultRegisterer)
	server := server.New(db, bundleManagerClient, codeIntelAPI)

	uploadResetter := resetter.UploadResetter{
		DB:            db,
		ResetInterval: resetInterval,
		Metrics:       resetterMetrics,
	}

	go server.Start()
	go uploadResetter.Run()
	go debugserver.Start()

	// Attempt to clean up after first shutdown signal
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP)
	<-signals

	go func() {
		// Insta-shutdown on a second signal
		<-signals
		os.Exit(0)
	}()

	server.Stop()
}

func mustInitializeDatabase() db.DB {
	postgresDSN := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.PostgresDSN; postgresDSN != newDSN {
			log.Fatalf("detected repository DSN change, restarting to take effect: %s", newDSN)
		}
	})

	db, err := db.New(postgresDSN)
	if err != nil {
		log.Fatalf("failed to initialize db store: %s", err)
	}

	return db
}
