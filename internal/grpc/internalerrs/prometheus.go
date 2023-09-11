package internalerrs

import (
	"context"
	"io"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var metricGRPCMethodStatus = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_grpc_method_status",
	Help: "Counts the number of gRPC methods that return a given status code, and whether a possible error is an go-grpc internal error.",
},
	[]string{
		"grpc_service",      // e.g. "gitserver.v1.GitserverService"
		"grpc_method",       // e.g. "Exec"
		"grpc_code",         // e.g. "NotFound"
		"is_internal_error", // e.g. "true"
	},
)

// PrometheusUnaryClientInterceptor returns a grpc.UnaryClientInterceptor that observes the result of
// the RPC and records it as a Prometheus metric ("src_grpc_method_status").
func PrometheusUnaryClientInterceptor(ctx context.Context, fullMethod string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	serviceName, methodName := grpcutil.SplitMethodName(fullMethod)

	err := invoker(ctx, fullMethod, req, reply, cc, opts...)
	doObservation(serviceName, methodName, err)
	return err
}

// PrometheusStreamClientInterceptor returns a grpc.StreamClientInterceptor that observes the result of
// the RPC and records it as a Prometheus metric ("src_grpc_method_status").
//
// If any errors are encountered during the stream, the first error is recorded. Otherwise, the
// final status of the stream is recorded.
func PrometheusStreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, fullMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	serviceName, methodName := grpcutil.SplitMethodName(fullMethod)

	s, err := streamer(ctx, desc, cc, fullMethod, opts...)
	if err != nil {
		doObservation(serviceName, methodName, err) // method failed to be invoked at all, record it
		return nil, err
	}

	return newPrometheusServerStream(s, serviceName, methodName), err
}

// newPrometheusServerStream wraps a grpc.ClientStream to observe the first error
// encountered during the stream, if any.
func newPrometheusServerStream(s grpc.ClientStream, serviceName, methodName string) grpc.ClientStream {
	// Design note: We only want a single observation for each RPC call: it either succeeds or fails
	// with a single error. This ensures we do not double-count RPCs in Prometheus metrics.
	//
	// For unary calls this is straightforward, but for streaming RPCs we need to make a compromise. We only
	// observe the first error (either sending or receiving) that occurs during the stream, instead of every
	// error that occurs during the stream's lifespan. While this approach swallows some errors, it keeps the
	// Prometheus metric count clean and non-duplicated. The logging interceptor handles surfacing all errors
	// that are encountered during a stream.
	var observeOnce sync.Once

	return &callBackClientStream{
		ClientStream: s,
		postMessageSend: func(_ any, err error) {
			if err != nil {
				observeOnce.Do(func() {
					doObservation(serviceName, methodName, err)
				})
			}
		},
		postMessageReceive: func(_ any, err error) {
			if err != nil {
				if err == io.EOF {
					// EOF signals end of stream, not an error. We handle this by setting err to nil, because
					// we want to treat the stream as successfully completed.
					err = nil
				}

				observeOnce.Do(func() {
					doObservation(serviceName, methodName, err)
				})
			}
		},
	}

}

func doObservation(serviceName, methodName string, rpcErr error) {
	if rpcErr == nil {
		// No error occurred, so we record a successful call.
		metricGRPCMethodStatus.WithLabelValues(serviceName, methodName, codes.OK.String(), "false").Inc()
		return
	}

	s, ok := massageIntoStatusErr(rpcErr)
	if !ok {
		// An error occurred, but it was not an error that has a status.Status implementation. We record this as an unknown error.
		metricGRPCMethodStatus.WithLabelValues(serviceName, methodName, codes.Unknown.String(), "false").Inc()
		return
	}

	if !probablyInternalGRPCError(s, allCheckers) {
		// An error occurred, but it was not an internal gRPC error. We record this as a non-internal error.
		metricGRPCMethodStatus.WithLabelValues(serviceName, methodName, s.Code().String(), "false").Inc()
		return
	}

	// An error occurred, and it looks like an internal gRPC error. We record this as an internal error.
	metricGRPCMethodStatus.WithLabelValues(serviceName, methodName, s.Code().String(), "true").Inc()
}
