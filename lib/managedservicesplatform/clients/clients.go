package clients

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// NewRecommendedGRPCDialOptions provides additional networking options tailored
// for connecting to typical MSP environments over gRPC (Cloud Run services
// behind Cloudflare).
func NewRecommendedGRPCDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
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
}
