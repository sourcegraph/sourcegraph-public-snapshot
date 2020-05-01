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
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
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

	bundleManagerClient := bundles.New(bundleManagerURL)

	dbMetrics := db.NewDBMetrics("precise_code_intel_api_server")
	dbMetrics.MustRegister(prometheus.DefaultRegisterer)

	db := db.NewObservedDB(
		mustInitializeDatabase(),
		log15.Root(),
		dbMetrics,
		trace.Tracer{Tracer: opentracing.GlobalTracer()},
	)

	codeIntelAPIMetrics := api.NewCodeIntelAPIMetrics("precise_code_intel_api_server")
	codeIntelAPIMetrics.MustRegister(prometheus.DefaultRegisterer)

	codeIntelAPI := api.NewObservedCodeIntelAPI(
		api.New(db, bundleManagerClient),
		log15.Root(),
		codeIntelAPIMetrics,
		trace.Tracer{Tracer: opentracing.GlobalTracer()},
	)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	serverInst := server.New(server.ServerOpts{
		Host:                host,
		Port:                3186,
		DB:                  db,
		BundleManagerClient: bundleManagerClient,
		CodeIntelAPI:        codeIntelAPI,
	})

	uploadResetterInst := resetter.NewUploadResetter(resetter.UploadResetterOpts{
		DB:            db,
		ResetInterval: resetInterval,
	})

	go serverInst.Start()
	go uploadResetterInst.Run()
	go debugserver.Start()
	waitForSignal()
}

func mustInitializeDatabase() db.DB {
	postgresDSN := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.PostgresDSN; postgresDSN != newDSN {
			log.Fatalf("Detected repository DSN change, restarting to take effect: %s", newDSN)
		}
	})

	db, err := db.New(postgresDSN)
	if err != nil {
		log.Fatalf("failed to initialize db store: %s", err)
	}

	return db
}

func waitForSignal() {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP)

	for i := 0; i < 2; i++ {
		<-signals
	}

	os.Exit(0)
}
