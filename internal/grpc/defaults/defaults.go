// package defaults exports a set of default options for gRPC servers
// and clients in Sourcegraph. It is a separate subpackage so that all
// packages that depend on the internal/grpc package do not need to
// depend on the large dependency tree of this package.
package defaults

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// DialOptions is a set of default dial options that should be used for all
// gRPC clients in Sourcegraph. The options can be extended with
// service-specific options.
func DialOptions() []grpc.DialOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.
	return []grpc.DialOption{
		grpc.WithChainStreamInterceptor(
			internalgrpc.StreamClientPropagator(actor.ActorPropagator{}),
			internalgrpc.StreamClientPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.StreamClientInterceptor(),
		),
		grpc.WithChainUnaryInterceptor(
			internalgrpc.UnaryClientPropagator(actor.ActorPropagator{}),
			internalgrpc.UnaryClientPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.UnaryClientInterceptor(),
		),
	}
}

// ServerOptions is a set of default server options that should be used for all
// gRPC servers in Sourcegrah. The options can be extended with
// service-specific options.
func ServerOptions() []grpc.ServerOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.
	return []grpc.ServerOption{
		grpc.ChainStreamInterceptor(
			internalgrpc.StreamServerPropagator(actor.ActorPropagator{}),
			internalgrpc.StreamServerPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			internalgrpc.UnaryServerPropagator(actor.ActorPropagator{}),
			internalgrpc.UnaryServerPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.UnaryServerInterceptor(),
		),
	}
}
