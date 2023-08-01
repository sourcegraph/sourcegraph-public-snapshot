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
	"github.com/sourcegraph/sourcegraph/internal/grpc/contextconv"
	"github.com/sourcegraph/sourcegraph/internal/grpc/messagesize"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/internalerrs"
	"github.com/sourcegraph/sourcegraph/internal/grpc/propagator"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
)

// Dial creates a client connection to the given target with the default options.
func Dial(addr string, logger log.Logger, additionalOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return DialContext(context.Background(), addr, logger, additionalOpts...)
}

// DialContext creates a client connection to the given target with the default options.
func DialContext(ctx context.Context, addr string, logger log.Logger, additionalOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, addr, DialOptions(logger, additionalOpts...)...)
}

// defaultGRPCMessageReceiveSizeBytes is the default message size that gRPCs servers and clients are allowed to process.
// This can be overridden by providing custom Server/Dial options.
const defaultGRPCMessageReceiveSizeBytes = 90 * 1024 * 1024 // 90 MB

// DialOptions is a set of default dial options that should be used for all
// gRPC clients in Sourcegraph, along with any additional client-specific options.
//
// **Note**: Do not append to this slice directly, instead provide extra options
// via "additionalOptions".
func DialOptions(logger log.Logger, additionalOptions ...grpc.DialOption) []grpc.DialOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.

	metrics := mustGetClientMetrics()

	out := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainStreamInterceptor(
			grpc_prometheus.StreamClientInterceptor(metrics),
			messagesize.StreamClientInterceptor,
			propagator.StreamClientPropagator(actor.ActorPropagator{}),
			propagator.StreamClientPropagator(policy.ShouldTracePropagator{}),
			propagator.StreamClientPropagator(requestclient.Propagator{}),
			otelgrpc.StreamClientInterceptor(),
			internalerrs.PrometheusStreamClientInterceptor,
			internalerrs.LoggingStreamClientInterceptor(logger),
			contextconv.StreamClientInterceptor,
		),
		grpc.WithChainUnaryInterceptor(
			grpc_prometheus.UnaryClientInterceptor(metrics),
			messagesize.UnaryClientInterceptor,
			propagator.UnaryClientPropagator(actor.ActorPropagator{}),
			propagator.UnaryClientPropagator(policy.ShouldTracePropagator{}),
			propagator.UnaryClientPropagator(requestclient.Propagator{}),
			otelgrpc.UnaryClientInterceptor(),
			internalerrs.PrometheusUnaryClientInterceptor,
			internalerrs.LoggingUnaryClientInterceptor(logger),
			contextconv.UnaryClientInterceptor,
		),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(defaultGRPCMessageReceiveSizeBytes)),
	}

	out = append(out, additionalOptions...)

	// Ensure that the message size options are set last, so they override any other
	// client-specific options that tweak the message size.
	//
	// The message size options are only provided if the environment variable is set. These options serve as an escape hatch, so they
	// take precedence over everything else with a uniform size setting that's easy to reason about.
	out = append(out, messagesize.MustGetClientMessageSizeFromEnv()...)

	return out
}

// NewServer creates a new *grpc.Server with the default options
func NewServer(logger log.Logger, additionalOpts ...grpc.ServerOption) *grpc.Server {
	s := grpc.NewServer(ServerOptions(logger, additionalOpts...)...)
	reflection.Register(s)
	return s
}

// ServerOptions is a set of default server options that should be used for all
// gRPC servers in Sourcegraph. along with any additional service-specific options.
//
// **Note**: Do not append to this slice directly, instead provide extra options
// via "additionalOptions".
func ServerOptions(logger log.Logger, additionalOptions ...grpc.ServerOption) []grpc.ServerOption {
	// Generate the options dynamically rather than using a static slice
	// because these options depend on some globals (tracer, trace sampling)
	// that are not initialized during init time.

	metrics := mustGetServerMetrics()

	out := []grpc.ServerOption{
		grpc.ChainStreamInterceptor(
			internalgrpc.NewStreamPanicCatcher(logger),
			internalerrs.LoggingStreamServerInterceptor(logger),
			grpc_prometheus.StreamServerInterceptor(metrics),
			messagesize.StreamServerInterceptor,
			propagator.StreamServerPropagator(requestclient.Propagator{}),
			propagator.StreamServerPropagator(actor.ActorPropagator{}),
			propagator.StreamServerPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.StreamServerInterceptor(),
			contextconv.StreamServerInterceptor,
		),
		grpc.ChainUnaryInterceptor(
			internalgrpc.NewUnaryPanicCatcher(logger),
			internalerrs.LoggingUnaryServerInterceptor(logger),
			grpc_prometheus.UnaryServerInterceptor(metrics),
			messagesize.UnaryServerInterceptor,
			propagator.UnaryServerPropagator(requestclient.Propagator{}),
			propagator.UnaryServerPropagator(actor.ActorPropagator{}),
			propagator.UnaryServerPropagator(policy.ShouldTracePropagator{}),
			otelgrpc.UnaryServerInterceptor(),
			contextconv.UnaryServerInterceptor,
		),
		grpc.MaxRecvMsgSize(defaultGRPCMessageReceiveSizeBytes),
	}

	out = append(out, additionalOptions...)

	// Ensure that the message size options are set last, so they override any other
	// server-specific options that tweak the message size.
	//
	// The message size options are only provided if the environment variable is set. These options serve as an escape hatch, so they
	// take precedence over everything else with a uniform size setting that's easy to reason about.
	out = append(out, messagesize.MustGetServerMessageSizeFromEnv()...)

	return out
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
