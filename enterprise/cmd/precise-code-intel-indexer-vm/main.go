package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/heartbeat"
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/index_manager"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/indexer"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/server"
	queue "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	//	tracer.Init() // TODO(efritz) - disabled as it requires internal API access

	var (
		frontendURL              = mustGet(rawFrontendURL, "PRECISE_CODE_INTEL_EXTERNAL_URL")
		frontendURLFromDocker    = mustGet(rawFrontendURLFromDocker, "PRECISE_CODE_INTEL_EXTERNAL_URL_FROM_DOCKER")
		internalProxyAuthToken   = mustGet(rawInternalProxyAuthToken, "PRECISE_CODE_INTEL_INTERNAL_PROXY_AUTH_TOKEN")
		indexerPollInterval      = mustParseInterval(rawIndexerPollInterval, "PRECISE_CODE_INTEL_INDEXER_POLL_INTERVAL")
		indexerHeartbeatInterval = mustParseInterval(rawIndexerHeartbeatInterval, "PRECISE_CODE_INTEL_INDEXER_HEARTBEAT_INTERVAL")
		numContainers            = mustParseInt(rawMaxContainers, "PRECISE_CODE_INTEL_MAXIMUM_CONTAINERS")
		firecrackerImage         = mustGet(rawFirecrackerImage, "PRECISE_CODE_INTEL_FIRECRACKER_IMAGE")
		useFirecracker           = mustParseBool(rawUseFirecracker, "PRECISE_CODE_INTEL_USE_FIRECRACKER")
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

	queueClient := queue.NewClient(
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
			FirecrackerImage:      firecrackerImage,
			UseFirecracker:        useFirecracker,
		},
	})

	go debugserver.Start()
	goroutine.MonitorBackgroundRoutines(server, indexer, heartbeater)
}
