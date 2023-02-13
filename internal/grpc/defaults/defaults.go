// package defaults exports a set of default options for gRPC servers
// and clients in Sourcegraph. It is a separate subpackage so that all
// packages that depend on the internal/grpc package do not need to
// depend on the large dependency tree of this package.
package defaults

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// DialOptions is a set of default dial options that should be used for all
// gRPC clients in Sourcegraph. The options can be extended with
// service-specific options.
func DialOptions(metrics *grpc_prometheus.ClientMetrics) []grpc.DialOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainStreamInterceptor(
			grpc_prometheus.StreamClientInterceptor(metrics),
			internalgrpc.StreamClientPropagator(actor.ActorPropagator{}),
			internalgrpc.StreamClientPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.StreamClientInterceptor(),
		),
		grpc.WithChainUnaryInterceptor(
			grpc_prometheus.UnaryClientInterceptor(metrics),
			internalgrpc.UnaryClientPropagator(actor.ActorPropagator{}),
			internalgrpc.UnaryClientPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.UnaryClientInterceptor(),
		),
	}
}

func RegisteredClientMetrics(serviceName string) *grpc_prometheus.ClientMetrics {
	return grpc_prometheus.NewRegisteredClientMetrics(prometheus.DefaultRegisterer,
		grpc_prometheus.WithClientCounterOptions(func(opts *prometheus.CounterOpts) {
			opts.Namespace = serviceName
		}),
		grpc_prometheus.WithClientHandlingTimeHistogram(), // record the overall request latency for a gRPC request
		grpc_prometheus.WithClientStreamRecvHistogram(),   // record how long it takes for a client to receive a message during a streaming RPC
		grpc_prometheus.WithClientStreamSendHistogram(),   // record how long it takes for a client to send a message during a streaming RPC
	)
}

func RegisteredServerMetrics(serviceName string) *grpc_prometheus.ServerMetrics {
	return grpc_prometheus.NewRegisteredServerMetrics(prometheus.DefaultRegisterer,
		grpc_prometheus.WithServerCounterOptions(func(opts *prometheus.CounterOpts) {
			opts.Namespace = serviceName
		}),
		grpc_prometheus.WithServerHandlingTimeHistogram(), // record the overall response latency for a gRPC request
	)
}

// ServerOptions is a set of default server options that should be used for all
// gRPC servers in Sourcegrah. The options can be extended with
// service-specific options.
func ServerOptions(logger log.Logger, metrics *grpc_prometheus.ServerMetrics) []grpc.ServerOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.
	return []grpc.ServerOption{
		grpc.ChainStreamInterceptor(
			internalgrpc.NewStreamPanicCatcher(logger),
			grpc_prometheus.StreamServerInterceptor(metrics),
			internalgrpc.StreamServerPropagator(actor.ActorPropagator{}),
			internalgrpc.StreamServerPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			internalgrpc.NewUnaryPanicCatcher(logger),
			grpc_prometheus.UnaryServerInterceptor(metrics),
			internalgrpc.UnaryServerPropagator(actor.ActorPropagator{}),
			internalgrpc.UnaryServerPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.UnaryServerInterceptor(),
		),
	}
}
