package v1

import (
	"context"
	"net/url"
	"time"

	"github.com/sourcegraph/log"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc/credentials/oauth"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Experimental: some additional networking options to account for some odd
// behaviour exhibited in Cloudflare when using gRPC.
var cloudRunDialOptions = []grpc.DialOption{
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		// Keep idle connections alive by pinging in this interval.
		// Default: Infinity.
		Time: 20 * time.Second,
		// Keepalive timeout duration.
		// Default: 20 seconds.
		Timeout: 10 * time.Second,
		// Ensure idle connections remain alive even if there is no traffic.
		// Default: False.
		PermitWithoutStream: true,
	}),
	// Ensure idle connections are not retained for a long time, to avoid
	// potential networking issues.
	// Default: 30 minutes
	grpc.WithIdleTimeout(1 * time.Minute),
}

type SimpleClient struct {
	TelemeteryGatewayServiceClient
	// Conn can be closed to ensure clean shutdown.
	Conn *grpc.ClientConn
}

// Dial establishes a connection to the Telemetry Gateway gRPC service with
// the given configuration. The oauth2.TokenSource should provide SAMS credentials,
// for example:
//
//	import (
//		sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
//		telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
//	)
//
//	func main() {
//		client, err := telemetrygatewayv1.Dial(ctx,
//			logger.Scoped("telemetrygateway"),
//			telemetryGatewayURL,
//			// Authenticate using SAMS client credentials
//			sams.ClientCredentialsTokenSource(
//				cfg.SAMSClientConfig.ConnConfig,
//				cfg.SAMSClientConfig.ClientID,
//				cfg.SAMSClientConfig.ClientSecret,
//				[]scopes.Scope{
//					scopes.ToScope(scopes.ServiceTelemetryGateway, "events", scopes.ActionWrite),
//				},
//			),
//		)
//		// ...
//	}
//
// Dial is intended for simple, standard production use cases. If you need
// to customize the way you connect to Telemetry Gateway, you should create your
// own dial setup.
func Dial(ctx context.Context, logger log.Logger, addr *url.URL, ts oauth2.TokenSource, dialOpts ...grpc.DialOption) (*SimpleClient, error) {
	insecureTarget := addr.Scheme != "https"
	if insecureTarget {
		return nil, errors.Newf("insecure target Telemetry Gateway %q", addr.String())
	}
	dialOpts = append(dialOpts, grpc.WithPerRPCCredentials(oauth.TokenSource{TokenSource: ts}))
	dialOpts = append(dialOpts, cloudRunDialOptions...)
	logger.Info("dialing Telemetry Gateway gRPC service",
		log.String("host", addr.Host),
		log.Bool("insecureTarget", insecureTarget))
	//lint:ignore SA1019 DialContext will be supported throughout 1.x
	conn, err := grpc.NewClient(addr.Host, dialOpts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to Telemetry Gateway gRPC service at %s", addr.String())
	}
	return &SimpleClient{
		TelemeteryGatewayServiceClient: NewTelemeteryGatewayServiceClient(conn),
		Conn:                           conn,
	}, nil
}
