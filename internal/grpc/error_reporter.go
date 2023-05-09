package grpc

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const envDisableGRPCInternalErrorReporting = "SRC_DISABLE_GRPC_INTERNAL_ERROR_REPORTING"

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

type ErrorReporter struct {
	log.Logger
}

func InternalErrorReporterUnaryClientIntereptor(logger log.Logger, serviceName string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)

		code := codes.OK.String()
		isInternalError

		if err == nil {
			metricGRPCMethodStatus.WithLabelValues(serviceName, method, codes.OK.String(), "false").Inc()
		} else {

			code = grpc.Code(err).String()
			metricGRPCMethodStatus.WithLabelValues(serviceName, method, code, "true").Inc()
			if !isInternalError(err) {
				logger.Error("grpc: internal error", log.String("method", method), log.Error(err))
			}
		}

		if err != nil {
			metricGRPCMethodStatus.WithLabelValues(serviceName, method, grpc.Code(err).String(), "true").Inc()
			if !isInternalError(err) {
				logger.Error("grpc: internal error", log.String("method", method), log.Error(err))
			}
		}

		if err != nil {

			metricGRPCMethodStatus.WithLabelValues(serviceName, method, grpc.Code(err).String(), "true").Inc()
			if !isInternalError(err) {
				logger.Error("grpc: internal error", log.String("method", method), log.Error(err))
			}
		}

		return err

	}
}
