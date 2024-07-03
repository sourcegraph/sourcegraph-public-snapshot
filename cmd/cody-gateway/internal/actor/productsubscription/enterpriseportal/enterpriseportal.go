package enterpriseportal

import (
	"context"
	"net/url"
	"time"

	"github.com/sourcegraph/log"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcoauth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Experimental: some additional networking options to account for some odd
// behaviour exhibited in Cloud Run.
var cloudRunDialOptions = []grpc.DialOption{
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		// Keep idle connections alive by pinging in this internval.
		// Default: Infinity.
		Time: 30 * time.Second,
		// Ensure idle connections remain alive even if there is no traffic.
		// Default: False.
		PermitWithoutStream: true,
	}),
	// Ensure idle connections are not retained for a long time, to avoid
	// potential networking issues.
	grpc.WithIdleTimeout(5 * time.Minute),
}

// Dial establishes a connection to the Enterprise Portal gRPC service with
// the given configuration. The oauth2.TokenSource should provide SAMS credentials.
func Dial(ctx context.Context, logger log.Logger, addr *url.URL, ts oauth2.TokenSource) (*grpc.ClientConn, error) {
	insecureTarget := addr.Scheme != "https"
	if insecureTarget && !env.InsecureDev {
		return nil, errors.New("insecure target Enterprise Portal used outside of dev mode")
	}
	creds := grpc.WithPerRPCCredentials(grpcoauth.TokenSource{TokenSource: ts})
	var opts []grpc.DialOption
	if insecureTarget {
		opts = defaults.DialOptions(logger, creds)
	} else {
		opts = defaults.ExternalDialOptions(logger,
			append([]grpc.DialOption{creds}, cloudRunDialOptions...)...)
	}
	logger.Info("dialing Enterprise Portal gRPC service",
		log.String("host", addr.Host),
		log.Bool("insecureTarget", insecureTarget))
	//lint:ignore SA1019 DialContext will be supported throughout 1.x
	conn, err := grpc.DialContext(ctx, addr.Host, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to Enterprise Portal gRPC service at %s", addr.String())
	}
	return conn, nil
}
