// Package defaults exports a set of default options for gRPC servers
// and clients in Sourcegraph. It is a separate subpackage so that all
// packages that depend on the internal/grpc package do not need to
// depend on the large dependency tree of this package.
package defaults

import (
	"context"
	"strings"
	"sync"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// Dial creates a client connection to the given target with the default options.
func Dial(addr string, additionalOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return DialContext(context.Background(), addr, additionalOpts...)
}

// DialContext creates a client connection to the given target with the default options.
func DialContext(ctx context.Context, addr string, additionalOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, addr, append(DialOptions(), additionalOpts...)...)
}

// DialOptions is a set of default dial options that should be used for all
// gRPC clients in Sourcegraph. The options can be extended with
// service-specific options.
func DialOptions() []grpc.DialOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.

	metrics := mustGetClientMetrics()

	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainStreamInterceptor(
			grpc_prometheus.StreamClientInterceptor(metrics),
			internalgrpc.StreamClientPropagator(actor.ActorPropagator{}),
			internalgrpc.StreamClientPropagator(policy.ShouldTracePropagator{}),
			otelStreamInterceptor,
		),
		grpc.WithChainUnaryInterceptor(
			grpc_prometheus.UnaryClientInterceptor(metrics),
			internalgrpc.UnaryClientPropagator(actor.ActorPropagator{}),
			internalgrpc.UnaryClientPropagator(policy.ShouldTracePropagator{}),
			otelUnaryInterceptor,
		),
	}
}

var (
	// Package-level variables because they are somewhat expensive to recreate every time
	otelStreamInterceptor = otelgrpc.StreamClientInterceptor()
	otelUnaryInterceptor  = otelgrpc.UnaryClientInterceptor()
)

// NewServer creates a new *grpc.Server with the default options
func NewServer(logger log.Logger, additionalOpts ...grpc.ServerOption) *grpc.Server {
	s := grpc.NewServer(append(ServerOptions(logger), additionalOpts...)...)
	reflection.Register(s)
	return s
}

// ServerOptions is a set of default server options that should be used for all
// gRPC servers in Sourcegraph. The options can be extended with
// service-specific options.
func ServerOptions(logger log.Logger) []grpc.ServerOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.

	metrics := mustGetServerMetrics()

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

var (
	clientMetricsOnce sync.Once
	clientMetrics     *grpc_prometheus.ClientMetrics

	serverMetricsOnce sync.Once
	serverMetrics     *grpc_prometheus.ServerMetrics
)

// mustGetClientMetrics returns a singleton instance of the client metrics
// that are shared across all gRPC clients that this process creates.
//
// This function panics if the metrics cannot be registered with the default
// Prometheus registry.
func mustGetClientMetrics() *grpc_prometheus.ClientMetrics {
	clientMetricsOnce.Do(func() {
		clientMetrics = grpc_prometheus.NewRegisteredClientMetrics(prometheus.DefaultRegisterer,
			grpc_prometheus.WithClientCounterOptions(setCounterNamespace),
			grpc_prometheus.WithClientHandlingTimeHistogram(setHistogramNamespace), // record the overall request latency for a gRPC request
			grpc_prometheus.WithClientStreamRecvHistogram(setHistogramNamespace),   // record how long it takes for a client to receive a message during a streaming RPC
			grpc_prometheus.WithClientStreamSendHistogram(setHistogramNamespace),   // record how long it takes for a client to send a message during a streaming RPC
		)
	})

	return clientMetrics
}

// mustGetServerMetrics returns a singleton instance of the server metrics
// that are shared across all gRPC servers that this process creates.
//
// This function panics if the metrics cannot be registered with the default
// Prometheus registry.
func mustGetServerMetrics() *grpc_prometheus.ServerMetrics {
	serverMetricsOnce.Do(func() {
		serverMetrics = grpc_prometheus.NewRegisteredServerMetrics(prometheus.DefaultRegisterer,
			grpc_prometheus.WithServerCounterOptions(setCounterNamespace),
			grpc_prometheus.WithServerHandlingTimeHistogram(setHistogramNamespace), // record the overall response latency for a gRPC request)
		)
	})

	return serverMetrics
}

// setCounterNamespace is prometheus option that sets the namespace for counter
// metrics to the current process name.
func setCounterNamespace(opts *prometheus.CounterOpts) {
	opts.Namespace = processNamePrometheus()
}

// setHistogramNamespace is prometheus option that sets the namespace for histogram
// metrics to the current process name.
func setHistogramNamespace(opts *prometheus.HistogramOpts) {
	opts.Namespace = processNamePrometheus()
}

// processNamePrometheus returns the name of the current binary (e.g. "frontend", "gitserver", "github_proxy"), with some
// additional normalization so that it can be used as a Prometheus namespace.
func processNamePrometheus() string {
	base := env.MyName
	base = strings.ReplaceAll(base, "-", "_")
	base = strings.ReplaceAll(base, ".", "_")

	return base
}
