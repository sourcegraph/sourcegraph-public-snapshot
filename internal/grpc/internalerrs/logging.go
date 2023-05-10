package internalerrs

import (
	"context"
	"io"
	"os"
	"strconv"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	logScope       = "gRPC.internal.error.reporter"
	logDescription = "logs gRPC errors that appear to come from the go-grpc implementation"

	envDisableGRPCInternalErrorLogging = "SRC_DISABLE_GRPC_INTERNAL_ERROR_LOGGING"
)

// LoggingUnaryClientInterceptor returns a grpc.UnaryClientInterceptor that logs
// errors that appear to come from the go-grpc implementation.
func LoggingUnaryClientInterceptor(l log.Logger) grpc.UnaryClientInterceptor {
	if !isLoggingEnabled() {
		// Just return the default invoker if logging is disabled.
		return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}

	logger := l.Scoped(logScope, logDescription)
	return func(ctx context.Context, fullMethod string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, fullMethod, req, reply, cc, opts...)
		if err != nil {
			serviceName, methodName := splitMethodName(fullMethod)
			doLog(logger, serviceName, methodName, err)
		}

		return err
	}
}

// LoggingStreamClientInterceptor returns a grpc.StreamClientInterceptor that logs
// errors that appear to come from the go-grpc implementation.
func LoggingStreamClientInterceptor(l log.Logger) grpc.StreamClientInterceptor {
	if !isLoggingEnabled() {
		// Just return the default streamer if logging is disabled.
		return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return streamer(ctx, desc, cc, method, opts...)
		}
	}

	logger := l.Scoped(logScope, logDescription)

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, fullMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		serviceName, methodName := splitMethodName(fullMethod)

		stream, err := streamer(ctx, desc, cc, fullMethod, opts...)
		if err != nil {
			doLog(logger, serviceName, methodName, err)
			return nil, err
		}

		stream = newLoggingClientStream(stream, logger, serviceName, methodName)
		return stream, nil
	}
}

func newLoggingClientStream(s grpc.ClientStream, logger log.Logger, serviceName, methodName string) *callBackClientStream {
	return &callBackClientStream{
		ClientStream: s,
		postMessageSend: func(err error) {
			if err != nil {
				doLog(logger, serviceName, methodName, err)
			}
		},
		postMessageReceive: func(err error) {
			if err != nil && err != io.EOF { // EOF is expected at the end of a stream, so no need to log an error
				doLog(logger, serviceName, methodName, err)
			}
		},
	}
}

func doLog(logger log.Logger, serviceName, methodName string, err error) {
	if err == nil {
		return
	}

	s, ok := status.FromError(err)
	if !ok {
		return
	}

	if !probablyInternalGRPCError(s) {
		return
	}

	logger.Error(s.Message(),
		log.String("grpcService", serviceName),
		log.String("grpcMethod", methodName),
		log.String("grpcCode", s.Code().String()))

}

// isLoggingEnabled returns true if logging of internal gRPC errors is enabled.
func isLoggingEnabled() bool {
	disabledStr, exists := os.LookupEnv(envDisableGRPCInternalErrorLogging)
	if !exists {
		return true
	}

	disabled, parseErr := strconv.ParseBool(disabledStr)
	if parseErr != nil {
		// If the environment variable is set to an invalid value, we consider logging to be enabled.
		return true
	}

	return !disabled
}
