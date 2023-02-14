// Package defaults exports a set of default options for gRPC servers
// and clients in Sourcegraph. It is a separate subpackage so that all
// packages that depend on the internal/grpc package do not need to
// depend on the large dependency tree of this package.
package defaults

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

var (
	// clientMetrics is set of metrics that are used to instrument all gRPC clients that this binary uses.
	clientMetrics = grpc_prometheus.NewRegisteredClientMetrics(prometheus.DefaultRegisterer,
		grpc_prometheus.WithClientCounterOptions(setCounterNamespace),
		grpc_prometheus.WithClientHandlingTimeHistogram(setHistogramNamespace), // record the overall request latency for a gRPC request
		grpc_prometheus.WithClientStreamRecvHistogram(setHistogramNamespace),   // record how long it takes for a client to receive a message during a streaming RPC
		grpc_prometheus.WithClientStreamSendHistogram(setHistogramNamespace),   // record how long it takes for a client to send a message during a streaming RPC
	)

	// serverMetrics is a set of metrics that are used to instrument all gRPC servers that this binary creates.
	serverMetrics = grpc_prometheus.NewRegisteredServerMetrics(prometheus.DefaultRegisterer,
		grpc_prometheus.WithServerCounterOptions(setCounterNamespace),
		grpc_prometheus.WithServerHandlingTimeHistogram(setHistogramNamespace), // record the overall response latency for a gRPC request
	)

	// prometheus option to set the namespace for counter metrics to the process name
	setCounterNamespace = func(opts *prometheus.CounterOpts) {
		opts.Namespace = processNamePrometheus()
	}
	// prometheus option to set the namespace for histogram to the process name
	setHistogramNamespace = func(opts *prometheus.HistogramOpts) {
		opts.Namespace = processNamePrometheus()
	}
)

// DialOptions is a set of default dial options that should be used for all
// gRPC clients in Sourcegraph. The options can be extended with
// service-specific options.
func DialOptions() []grpc.DialOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainStreamInterceptor(
			grpc_prometheus.StreamClientInterceptor(clientMetrics),
			internalgrpc.StreamClientPropagator(actor.ActorPropagator{}),
			internalgrpc.StreamClientPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.StreamClientInterceptor(),
		),
		grpc.WithChainUnaryInterceptor(
			grpc_prometheus.UnaryClientInterceptor(clientMetrics),
			internalgrpc.UnaryClientPropagator(actor.ActorPropagator{}),
			internalgrpc.UnaryClientPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.UnaryClientInterceptor(),
		),
	}
}

// ServerOptions is a set of default server options that should be used for all
// gRPC servers in Sourcegraph. The options can be extended with
// service-specific options.
func ServerOptions(logger log.Logger) []grpc.ServerOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.
	return []grpc.ServerOption{
		grpc.ChainStreamInterceptor(
			internalgrpc.NewStreamPanicCatcher(logger),
			grpc_prometheus.StreamServerInterceptor(serverMetrics),
			internalgrpc.StreamServerPropagator(actor.ActorPropagator{}),
			internalgrpc.StreamServerPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			internalgrpc.NewUnaryPanicCatcher(logger),
			grpc_prometheus.UnaryServerInterceptor(serverMetrics),
			internalgrpc.UnaryServerPropagator(actor.ActorPropagator{}),
			internalgrpc.UnaryServerPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.UnaryServerInterceptor(),
		),
	}
}

// processNamePrometheus returns the name of the current binary (e.g. "frontend", "gitserver", "github_proxy"), with some
// additional normalization so that it can be used as a Prometheus namespace.
func processNamePrometheus() string {
	base := env.MyName
	base = strings.ReplaceAll(base, "-", "_")
	base = strings.ReplaceAll(base, ".", "_")

	return base
}
