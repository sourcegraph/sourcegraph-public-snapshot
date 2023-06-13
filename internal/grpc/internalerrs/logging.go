package internalerrs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

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

	envLogNonUTF8ProtobufMessages        = env.MustGetBool("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_NON_UTF8_PROTOBUF_MESSAGES_ENABLED", false, "Enables logging of non-UTF-8 protobuf messages")
	envLogNonUTF8ProtobufMessagesMaxSize = env.MustGetInt("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_NON_UTF8_PROTOBUF_MESSAGES_MAX_SIZE_BYTES", 1024, "Maximum size of non-UTF-8 protobuf messages to log before truncation, in bytes. Negative values disable truncation.")
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
			// Note: This is a bit hacky, we provide a nil message here since the message isn't available
			// until after the stream is created.
			//
			// This is fine since the error is already available, and the non-utf8 string check is robust against nil messages.
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

		postMessageSend: func(m any, err error) {
			if err != nil {
				doLog(logger, serviceName, methodName, m, err)
			}
		},

		postMessageReceive: func(m any, err error) {
			if err != nil && err != io.EOF { // EOF is expected at the end of a stream, so no need to log an error
				doLog(logger, serviceName, methodName, m, err)
			}
		},
	}
}

func doLog(logger log.Logger, serviceName, methodName string, payload any, err error) {
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

	allFields := []log.Field{
		log.String("grpcService", serviceName),
		log.String("grpcMethod", methodName),
		log.String("grpcCode", s.Code().String()),
	}

	if envLogStackTracesEnabled {
		allFields = append(allFields, log.String("errWithStack", fmt.Sprintf("%+v", err)))
	}

	if isNonUTF8StringError(s) {
		if m, ok := payload.(proto.Message); ok {
			allFields = append(
				allFields,
				additionalNonUTF8StringDebugFields(m, envLogNonUTF8ProtobufMessages, envLogNonUTF8ProtobufMessagesMaxSize)...,
			)
		}
	}

	logger.Error(s.Message(), allFields...)
}

// additionalNonUTF8StringDebugFields returns additional log fields that should be included when logging a non-UTF8 string error.
//
// By default, this includes the names of all fields that contain non-UTF8 strings.
// If shouldLogMessageJSON is true, then the JSON representation of the message is also included.
// The maxMessageSizeLogBytes parameter controls the maximum size of the message that will be logged, after which it will be truncated. Negative values disable truncation.
func additionalNonUTF8StringDebugFields(message proto.Message, shouldLogMessageJSON bool, maxMessageLogSizeBytes int) []log.Field {
	var allFields []log.Field

	// Add the names of all protobuf fields that contain non-UTF-8 strings to the log.

	badFields, err := findNonUTF8StringFields(message)
	if err != nil {
		allFields = append(allFields, log.Error(errors.Wrapf(err, "failed to find non-UTF8 string allFields")))
		return allFields
	}

	allFields = append(allFields, log.Strings("nonUTF8StringFields", badFields))

	// Add the JSON representation of the message to the log.

	if !shouldLogMessageJSON {
		return allFields
	}

	// Note: we can't use the protojson library here since it doesn't support messages with non-UTF8 strings.
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		allFields = append(allFields, log.Error(errors.Wrapf(err, "failed to marshal protobuf message to bytes")))
		return allFields
	}

	if maxMessageLogSizeBytes < 0 { // If truncation is disabled, set the max size to the full message size.
		maxMessageLogSizeBytes = len(jsonBytes)
	}

	bytesToTruncate := len(jsonBytes) - maxMessageLogSizeBytes
	if bytesToTruncate > 0 {
		jsonBytes = jsonBytes[:maxMessageLogSizeBytes]
		jsonBytes = append(jsonBytes, []byte(fmt.Sprintf("...(truncated %d bytes)", bytesToTruncate))...)
	}

	allFields = append(allFields, log.String("messageJSON", string(jsonBytes)))
	return allFields
}

func isNonUTF8StringError(s *status.Status) bool {
	if s.Code() != codes.Internal {
		return false
	}

	return strings.Contains(s.Message(), "string field contains invalid UTF-8")
}
