package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/heartbeat"
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/index_manager"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/indexer"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/server"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
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
		frontendURL              = mustGet(rawFrontendURL, "PRECISE_CODE_INTEL_EXTERNAL_URL")
		frontendURLFromDocker    = mustGet(rawFrontendURLFromDocker, "PRECISE_CODE_INTEL_EXTERNAL_URL_FROM_DOCKER")
		internalProxyAuthToken   = mustGet(rawInternalProxyAuthToken, "PRECISE_CODE_INTEL_INTERNAL_PROXY_AUTH_TOKEN")
		indexerPollInterval      = mustParseInterval(rawIndexerPollInterval, "PRECISE_CODE_INTEL_INDEXER_POLL_INTERVAL")
		indexerHeartbeatInterval = mustParseInterval(rawIndexerHeartbeatInterval, "PRECISE_CODE_INTEL_INDEXER_HEARTBEAT_INTERVAL")
		numContainers            = mustParseInt(rawMaxContainers, "PRECISE_CODE_INTEL_MAXIMUM_CONTAINERS")
	)

	if frontendURLFromDocker == "" {
		frontendURLFromDocker = frontendURL
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	indexerName := uuid.New().String()

	queueClient := client.NewClient(
		indexerName,
		frontendURL,
		internalProxyAuthToken,
	)
	indexManager := indexmanager.New()
	server := server.New()
	heartbeater := heartbeat.NewHeartbeater(context.Background(), queueClient, indexManager, heartbeat.HeartbeaterOptions{
		Interval: indexerHeartbeatInterval,
	})
	indexerMetrics := indexer.NewIndexerMetrics(observationContext)
	indexer := indexer.NewIndexer(context.Background(), queueClient, indexManager, indexer.IndexerOptions{
		NumIndexers: numContainers,
		Interval:    indexerPollInterval,
		Metrics:     indexerMetrics,
		HandlerOptions: indexer.HandlerOptions{
			FrontendURL:           frontendURL,
			FrontendURLFromDocker: frontendURLFromDocker,
			AuthToken:             internalProxyAuthToken,
		},
	})

	go server.Start()
	go indexer.Start()
	go debugserver.Start()
	go heartbeater.Start()

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP)
	<-signals

	go func() {
		// Insta-shutdown on a second signal
		<-signals
		os.Exit(0)
	}()

	server.Stop()
	indexer.Stop()
	heartbeater.Stop()
}
