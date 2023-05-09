package internalerrs

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var metricGRPCMethodStatus = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_grpc_method_status",
},
	[]string{
		"grpc_service",      // e.g. "gitserver.v1.GitserverService"
		"grpc_method",       // e.g. "Exec"
		"grpc_code",         // e.g. "NotFound"
		"is_internal_error", // e.g. "true"
	},
)

func PrometheusUnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	doObservation(method, err)
	return err
}

func doObservation(fullMethod string, rpcErr error) {
	serviceName, methodName := splitMethodName(fullMethod)
	if rpcErr == nil {
		metricGRPCMethodStatus.WithLabelValues(serviceName, methodName, codes.OK.String(), "false").Inc()
		return
	}

	s, ok := status.FromError(rpcErr)
	if !ok {
		metricGRPCMethodStatus.WithLabelValues(serviceName, methodName, codes.Unknown.String(), "false").Inc()
		return
	}

	if !probablyInternalGRPCError(s) {
		metricGRPCMethodStatus.WithLabelValues(serviceName, methodName, s.Code().String(), "false").Inc()
		return
	}

	metricGRPCMethodStatus.WithLabelValues(serviceName, methodName, s.Code().String(), "true").Inc()
	return
}
