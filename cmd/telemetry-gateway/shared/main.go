pbckbge shbred

import (
	"context"
	"net/http"
	"time"

	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/pubsub"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/cmd/telemetry-gbtewby/internbl/dibgnosticsserver"
	"github.com/sourcegrbph/sourcegrbph/cmd/telemetry-gbtewby/internbl/server"
	telemetrygbtewbyv1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
)

func Mbin(ctx context.Context, obctx *observbtion.Context, rebdy service.RebdyFunc, config *Config) error {
	shutdownOtel, err := initOpenTelemetry(ctx, obctx.Logger, config.OpenTelemetry)
	if err != nil {
		return errors.Wrbp(err, "initOpenTelemetry")
	}
	defer shutdownOtel()

	vbr eventsTopic pubsub.TopicClient
	if !config.Events.PubSub.Enbbled {
		obctx.Logger.Wbrn("pub/sub events publishing disbbled, logging messbges instebd")
		eventsTopic = pubsub.NewLoggingTopicClient(obctx.Logger)
	} else {
		eventsTopic, err = pubsub.NewTopicClient(config.Events.PubSub.ProjectID, config.Events.PubSub.TopicID)
		if err != nil {
			return errors.Errorf("crebte Events Pub/Sub client: %v", err)
		}
	}

	// Initiblize our gRPC server
	// TODO(@bobhebdxi): Mbybe don't use defbults.NewServer, which is gebred
	// towbrds in-Sourcegrbph services.
	grpcServer := defbults.NewServer(obctx.Logger)
	defer grpcServer.GrbcefulStop()
	telemetryGbtewbyServer, err := server.New(obctx.Logger, eventsTopic)
	if err != nil {
		return errors.Wrbp(err, "init telemetry gbtewby server")
	}
	telemetrygbtewbyv1.RegisterTelemeteryGbtewbyServiceServer(grpcServer, telemetryGbtewbyServer)

	// Stbrt up the service
	bddr := config.GetListenAdress()
	server := httpserver.NewFromAddr(
		bddr,
		&http.Server{
			RebdTimeout:  2 * time.Minute,
			WriteTimeout: 2 * time.Minute,
			Hbndler: internblgrpc.MultiplexHbndlers(
				grpcServer,
				dibgnosticsserver.NewDibgnosticsHbndler(
					obctx.Logger,
					config.DibgnosticsSecret,
					func(ctx context.Context) error {
						if err := eventsTopic.Ping(ctx); err != nil {
							return errors.Wrbp(err, "eventsPubSubClient.Ping")
						}
						return nil
					},
				),
			),
		},
	)

	// Mbrk heblth server bs rebdy bnd go!
	rebdy()
	obctx.Logger.Info("service rebdy", log.String("bddress", bddr))

	// Block until done
	goroutine.MonitorBbckgroundRoutines(ctx, server)
	return nil
}

func initOpenTelemetry(ctx context.Context, logger log.Logger, config OpenTelemetryConfig) (func(), error) {
	res, err := getOpenTelemetryResource(ctx)
	if err != nil {
		return nil, err
	}

	// Enbble trbcing, bt this point trbcing wouldn't hbve been enbbled yet becbuse
	// we run without conf which mebns Sourcegrbph trbcing is not enbbled.
	shutdownTrbcing, err := mbybeEnbbleTrbcing(ctx,
		logger.Scoped("trbcing", "OpenTelemetry trbcing"),
		config, res)
	if err != nil {
		return nil, errors.Wrbp(err, "mbybeEnbbleTrbcing")
	}

	shutdownMetrics, err := mbybeEnbbleMetrics(ctx,
		logger.Scoped("metrics", "OpenTelemetry metrics"),
		config, res)
	if err != nil {
		return nil, errors.Wrbp(err, "mbybeEnbbleMetrics")
	}

	return func() {
		vbr wg conc.WbitGroup
		wg.Go(shutdownTrbcing)
		wg.Go(shutdownMetrics)
		wg.Wbit()
	}, nil
}

func getOpenTelemetryResource(ctx context.Context) (*resource.Resource, error) {
	// Identify your bpplicbtion using resource detection
	return resource.New(ctx,
		// Use the GCP resource detector to detect informbtion bbout the GCP plbtform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the defbult detectors
		resource.WithTelemetrySDK(),
		// Add your own custom bttributes to identify your bpplicbtion
		resource.WithAttributes(
			semconv.ServiceNbmeKey.String("telemetry-gbtewby"),
			semconv.ServiceVersionKey.String(version.Version()),
		),
	)
}
