package shared

import (
	"context"
	"net/http"
	"time"

	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/diagnosticsserver"
	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/server"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	shutdownOtel, err := initOpenTelemetry(ctx, obctx.Logger, config.OpenTelemetry)
	if err != nil {
		return errors.Wrap(err, "initOpenTelemetry")
	}
	defer shutdownOtel()

	var eventsTopic pubsub.TopicClient
	if !config.Events.PubSub.Enabled {
		obctx.Logger.Warn("pub/sub events publishing disabled, logging messages instead")
		eventsTopic = pubsub.NewLoggingTopicClient(obctx.Logger)
	} else {
		eventsTopic, err = pubsub.NewTopicClient(config.Events.PubSub.ProjectID, config.Events.PubSub.TopicID)
		if err != nil {
			return errors.Errorf("create Events Pub/Sub client: %v", err)
		}
	}

	// Initialize our gRPC server
	// TODO(@bobheadxi): Maybe don't use defaults.NewServer, which is geared
	// towards in-Sourcegraph services.
	grpcServer := defaults.NewServer(obctx.Logger)
	defer grpcServer.GracefulStop()
	telemetryGatewayServer, err := server.New(obctx.Logger, eventsTopic)
	if err != nil {
		return errors.Wrap(err, "init telemetry gateway server")
	}
	telemetrygatewayv1.RegisterTelemeteryGatewayServiceServer(grpcServer, telemetryGatewayServer)

	// Start up the service
	addr := config.GetListenAdress()
	server := httpserver.NewFromAddr(
		addr,
		&http.Server{
			ReadTimeout:  2 * time.Minute,
			WriteTimeout: 2 * time.Minute,
			Handler: internalgrpc.MultiplexHandlers(
				grpcServer,
				diagnosticsserver.NewDiagnosticsHandler(
					obctx.Logger,
					config.DiagnosticsSecret,
					func(ctx context.Context) error {
						if err := eventsTopic.Ping(ctx); err != nil {
							return errors.Wrap(err, "eventsPubSubClient.Ping")
						}
						return nil
					},
				),
			),
		},
	)

	// Mark health server as ready and go!
	ready()
	obctx.Logger.Info("service ready", log.String("address", addr))

	// Block until done
	goroutine.MonitorBackgroundRoutines(ctx, server)
	return nil
}

func initOpenTelemetry(ctx context.Context, logger log.Logger, config OpenTelemetryConfig) (func(), error) {
	res, err := getOpenTelemetryResource(ctx)
	if err != nil {
		return nil, err
	}

	// Enable tracing, at this point tracing wouldn't have been enabled yet because
	// we run without conf which means Sourcegraph tracing is not enabled.
	shutdownTracing, err := maybeEnableTracing(ctx,
		logger.Scoped("tracing"),
		config, res)
	if err != nil {
		return nil, errors.Wrap(err, "maybeEnableTracing")
	}

	shutdownMetrics, err := maybeEnableMetrics(ctx,
		logger.Scoped("metrics"),
		config, res)
	if err != nil {
		return nil, errors.Wrap(err, "maybeEnableMetrics")
	}

	return func() {
		var wg conc.WaitGroup
		wg.Go(shutdownTracing)
		wg.Go(shutdownMetrics)
		wg.Wait()
	}, nil
}

func getOpenTelemetryResource(ctx context.Context) (*resource.Resource, error) {
	// Identify your application using resource detection
	return resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String("telemetry-gateway"),
			semconv.ServiceVersionKey.String(version.Version()),
		),
	)
}
