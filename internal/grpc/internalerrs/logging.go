package internalerrs

import (
	"context"
	"fmt"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"io"
	"strings"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	logScope       = "gRPC.internal.error.reporter"
	logDescription = "logs gRPC errors that appear to come from the go-grpc implementation"

	envLoggingEnabled        = env.MustGetBool("SRC_GRPC_INTERNAL_ERROR_LOGGING_ENABLED", true, "Enables logging of gRPC internal errors")
	envLogStackTracesEnabled = env.MustGetBool("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_STACK_TRACES", false, "Enables including stack traces in logs of gRPC internal errors")

	envLogNonUTF8ProtobufMessages        = env.MustGetBool("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_NON_UTF8_PROTOBUF_MESSAGES", true, "Enables logging of non-UTF-8 protobuf messages")
	envLogNonUTF8ProtobufMessagesMaxSize = env.MustGetInt("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_NON_UTF8_PROTOBUF_MESSAGES_MAX_SIZE", 1024, "Maximum size of non-UTF-8 protobuf messages to log, in bytes")
)

// LoggingUnaryClientInterceptor returns a grpc.UnaryClientInterceptor that logs
// errors that appear to come from the go-grpc implementation.
func LoggingUnaryClientInterceptor(l log.Logger) grpc.UnaryClientInterceptor {
	if !envLoggingEnabled {
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
			doLog(logger, serviceName, methodName, req, err)
		}

		return err
	}
}

// LoggingStreamClientInterceptor returns a grpc.StreamClientInterceptor that logs
// errors that appear to come from the go-grpc implementation.
func LoggingStreamClientInterceptor(l log.Logger) grpc.StreamClientInterceptor {
	if !envLoggingEnabled {
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
			doLog(logger, serviceName, methodName, nil, err)
			return nil, err
		}

		stream = newLoggingClientStream(stream, logger, serviceName, methodName)
		return stream, nil
	}
}

func newLoggingClientStream(s grpc.ClientStream, logger log.Logger, serviceName, methodName string) *callBackClientStream {
	return &callBackClientStream{
		ClientStream: s,
		postMessageSend: func(m interface{}, err error) {
			if err != nil {
				doLog(logger, serviceName, methodName, m, err)
			}
		},
		postMessageReceive: func(m interface{}, err error) {
			if err != nil && err != io.EOF { // EOF is expected at the end of a stream, so no need to log an error
				doLog(logger, serviceName, methodName, m, err)
			}
		},
	}
}

func doLog(logger log.Logger, serviceName, methodName string, message interface{}, err error) {
	if err == nil {
		return
	}

	s, ok := status.FromError(err)
	if !ok {
		return
	}

	if !probablyInternalGRPCError(s, allCheckers) {
		return
	}

	fields := []log.Field{
		log.String("grpcService", serviceName),
		log.String("grpcMethod", methodName),
		log.String("grpcCode", s.Code().String()),
	}

	if envLogStackTracesEnabled {
		fields = append(fields, log.String("errWithStack", fmt.Sprintf("%+v", err)))
	}

	if isNonUTF8StringError(s) {
		if m, ok := message.(proto.Message); ok {
			badStrings, err := findNonUTF8StringFields(m)
			if err != nil {
				fields = append(fields, log.Error(errors.Wrapf(err, "failed to find non-UTF8 string fields")))
			} else {
				fields = append(fields, log.Strings("nonUTF8StringFields", badStrings))
			}

			if envLogNonUTF8ProtobufMessages {
				messageString := prototext.MarshalOptions{AllowPartial: true}.Format(m)
				messageString = truncate(messageString, envLogNonUTF8ProtobufMessagesMaxSize)

				fields = append(fields, log.String("message", messageString))
			}
		}
	}

	logger.Error(s.Message(), fields...)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen] + "...(truncated)"
}

func isNonUTF8StringError(s *status.Status) bool {
	if s.Code() != codes.Internal {
		return false
	}

	return strings.Contains(s.Message(), "string field contains invalid UTF-8")
}
