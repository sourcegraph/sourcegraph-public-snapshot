package shared

import (
	"context"
	"net/http"
	"time"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/internal/version"

	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/server"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/internal/telemetrygateway/v1"
)

type Service struct{}

var _ runtime.Service[Config] = (*Service)(nil)

func (Service) Name() string    { return "telemetry-gateway" }
func (Service) Version() string { return version.Version() }

func (Service) Initialize(ctx context.Context, logger log.Logger, contract runtime.Contract, config Config) (background.CombinedRoutine, error) {
	// We use Sourcegraph tracing code, so explicitly configure a trace policy
	policy.SetTracePolicy(policy.TraceAll)

	// Prepare pubsub client
	var err error
	var eventsTopic pubsub.TopicClient
	if !config.Events.PubSub.Enabled {
		logger.Warn("pub/sub events publishing disabled, logging messages instead")
		eventsTopic = pubsub.NewLoggingTopicClient(logger)
	} else {
		eventsTopic, err = pubsub.NewTopicClient(*config.Events.PubSub.ProjectID, *config.Events.PubSub.TopicID)
		if err != nil {
			return nil, errors.Errorf("create Events Pub/Sub client: %v", err)
		}
	}

	// Initialize our gRPC server
	// TODO(@bobheadxi): Maybe don't use defaults.NewServer, which is geared
	// towards in-Sourcegraph services.
	grpcServer := defaults.NewServer(logger)
	telemetryGatewayServer, err := server.New(logger, eventsTopic)
	if err != nil {
		return nil, errors.Wrap(err, "init telemetry gateway server")
	}
	telemetrygatewayv1.RegisterTelemeteryGatewayServiceServer(grpcServer, telemetryGatewayServer)

	// Set up diagnostics endpoints
	diagnosticsServer := http.NewServeMux()
	contract.RegisterDiagnosticsHandlers(diagnosticsServer, &serviceStatus{
		eventsTopic: eventsTopic,
	})
	if !contract.MSP {
		// Requires GRPC_WEB_UI_ENABLED to be set to enable - only use in local
		// development!
		grpcUI := debugserver.NewGRPCWebUIEndpoint("telemetry-gateway", config.GetListenAdress())
		diagnosticsServer.Handle(grpcUI.Path, grpcUI.Handler)
	}

	return background.CombinedRoutine{
		httpserver.NewFromAddr(
			config.GetListenAdress(),
			&http.Server{
				ReadTimeout:  2 * time.Minute,
				WriteTimeout: 2 * time.Minute,
				Handler: internalgrpc.MultiplexHandlers(
					grpcServer,
					diagnosticsServer,
				),
			},
		),
		grpcServerStopper{server: grpcServer},
	}, nil
}

type serviceStatus struct {
	eventsTopic pubsub.TopicClient
}

func (s *serviceStatus) Healthy(ctx context.Context) error {
	if err := s.eventsTopic.Ping(ctx); err != nil {
		return errors.Wrap(err, "eventsPubSubClient.Ping")
	}
	return nil
}

type grpcServerStopper struct{ server *grpc.Server }

func (g grpcServerStopper) Start() {} // nothing to do
func (g grpcServerStopper) Stop()  { g.server.GracefulStop() }
