package enterpriseportal

import (
	"context"
	"net/url"

	"github.com/sourcegraph/log"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcoauth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/clients"
)

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
			append([]grpc.DialOption{creds}, clients.NewRecommendedGRPCDialOptions()...)...)
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
