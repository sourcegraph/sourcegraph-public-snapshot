package internalerrs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/errors"

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
	logger = logger.Scoped("unaryMethod", "errors that originated from a unary method")

	return func(ctx context.Context, fullMethod string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, fullMethod, req, reply, cc, opts...)
		if err != nil {
			serviceName, methodName := splitMethodName(fullMethod)

			var initialRequest proto.Message
			if m, ok := req.(proto.Message); ok {
				initialRequest = m
			}

			doLog(logger, serviceName, methodName, &initialRequest, req, err)
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
	logger = logger.Scoped("streamingMethod", "errors that originated from a streaming method")

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, fullMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		serviceName, methodName := splitMethodName(fullMethod)

		stream, err := streamer(ctx, desc, cc, fullMethod, opts...)
		if err != nil {
			// Note: This is a bit hacky, we provide nil initial and payload messages here since the message isn't available
			// until after the stream is created.
			//
			// This is fine since the error is already available, and the non-utf8 string check is robust against nil messages.
			logger := logger.Scoped("postInit", "errors that occurred after stream initialization, but before the first message was sent")
			doLog(logger, serviceName, methodName, nil, nil, err)
			return nil, err
		}

		stream = newLoggingClientStream(stream, logger, serviceName, methodName)
		return stream, nil
	}
}

// LoggingUnaryServerInterceptor returns a grpc.UnaryServerInterceptor that logs
// errors that appear to come from the go-grpc implementation.
func LoggingUnaryServerInterceptor(l log.Logger) grpc.UnaryServerInterceptor {
	if !envLoggingEnabled {
		// Just return the default handler if logging is disabled.
		return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			return handler(ctx, req)
		}
	}

	logger := l.Scoped(logScope, logDescription)
	logger = logger.Scoped("unaryMethod", "errors that originated from a unary method")

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		response, err := handler(ctx, req)
		if err != nil {
			serviceName, methodName := splitMethodName(info.FullMethod)

			var initialRequest proto.Message
			if m, ok := req.(proto.Message); ok {
				initialRequest = m
			}

			doLog(logger, serviceName, methodName, &initialRequest, response, err)
		}

		return response, err
	}
}

// LoggingStreamServerInterceptor returns a grpc.StreamServerInterceptor that logs
// errors that appear to come from the go-grpc implementation.
func LoggingStreamServerInterceptor(l log.Logger) grpc.StreamServerInterceptor {
	if !envLoggingEnabled {
		// Just return the default handler if logging is disabled.
		return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, ss)
		}
	}

	logger := l.Scoped(logScope, logDescription)
	logger = logger.Scoped("streamingMethod", "errors that originated from a streaming method")

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		serviceName, methodName := splitMethodName(info.FullMethod)

		stream := newLoggingServerStream(ss, logger, serviceName, methodName)
		return handler(srv, stream)
	}
}

func newLoggingServerStream(s grpc.ServerStream, logger log.Logger, serviceName, methodName string) grpc.ServerStream {
	sendLogger := logger.Scoped("postMessageSend", "errors that occurred after sending a message")
	receiveLogger := logger.Scoped("postMessageReceive", "errors that occurred after receiving a message")

	requestSaver := requestSavingServerStream{ServerStream: s}

	return &callBackServerStream{
		ServerStream: &requestSaver,

		postMessageSend: func(m any, err error) {
			if err != nil {
				doLog(sendLogger, serviceName, methodName, requestSaver.InitialRequest(), m, err)
			}
		},

		postMessageReceive: func(m any, err error) {
			if err != nil && err != io.EOF { // EOF is expected at the end of a stream, so no need to log an error
				doLog(receiveLogger, serviceName, methodName, requestSaver.InitialRequest(), m, err)
			}
		},
	}
}

func newLoggingClientStream(s grpc.ClientStream, logger log.Logger, serviceName, methodName string) grpc.ClientStream {
	sendLogger := logger.Scoped("postMessageSend", "errors that occurred after sending a message")
	receiveLogger := logger.Scoped("postMessageReceive", "errors that occurred after receiving a message")

	requestSaver := requestSavingClientStream{ClientStream: s}

	return &callBackClientStream{
		ClientStream: &requestSaver,

		postMessageSend: func(m any, err error) {
			if err != nil {
				doLog(sendLogger, serviceName, methodName, requestSaver.InitialRequest(), m, err)
			}
		},

		postMessageReceive: func(m any, err error) {
			if err != nil && err != io.EOF { // EOF is expected at the end of a stream, so no need to log an error
				doLog(receiveLogger, serviceName, methodName, requestSaver.InitialRequest(), m, err)
			}
		},
	}
}

func doLog(logger log.Logger, serviceName, methodName string, initialRequest *proto.Message, payload any, err error) {
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
		m, ok := payload.(proto.Message)
		if ok {
			allFields = append(
				allFields,
				additionalNonUTF8StringDebugFields(initialRequest, m, envLogNonUTF8ProtobufMessages, envLogNonUTF8ProtobufMessagesMaxSize)...,
			)
		}
	}

	logger.Error(s.Message(), allFields...)
}

// additionalNonUTF8StringDebugFields returns additional log fields that should be included when logging a non-UTF8 string error.
//
// By default, this includes the names of all fields that contain non-UTF8 strings.
// If shouldLogMessageJSON is true, then the JSON representations for the initial request message and latest message are also included.
// The maxMessageSizeLogBytes parameter controls the maximum size of the messages that will be logged, after which they will be truncated. Negative values disable truncation.
func additionalNonUTF8StringDebugFields(firstMessage *proto.Message, latestMessage proto.Message, shouldLogMessageJSON bool, maxMessageLogSizeBytes int) []log.Field {
	var allFields []log.Field

	// Add the names of all protobuf fields that contain non-UTF-8 strings to the log.

	badFields, err := findNonUTF8StringFields(latestMessage)
	if err != nil {
		allFields = append(allFields, log.Error(errors.Wrapf(err, "failed to find non-UTF8 string fields in protobuf message")))
		return allFields
	}

	allFields = append(allFields, log.Strings("nonUTF8StringFields", badFields))

	// Add the JSON representation of the message to the log.

	if !shouldLogMessageJSON {
		return allFields
	}

	// Note: we can't use the protojson library here since it doesn't support messages with non-UTF8 strings.
	jsonBytes, err := json.Marshal(latestMessage)
	if err != nil {
		allFields = append(allFields, log.Error(errors.Wrapf(err, "failed to marshal latest protobuf message to bytes")))
		return allFields
	}

	s := truncate(string(jsonBytes), maxMessageLogSizeBytes)
	allFields = append(allFields, log.String("messageJSON", s))

	// Log the JSON representation of the initial request message, if available for debugging purposes.

	if firstMessage == nil {
		return allFields
	}

	jsonBytes, err = json.Marshal(*firstMessage)
	if err != nil {
		allFields = append(allFields, log.Error(errors.Wrapf(err, "failed to marshal initial request protobuf message to bytes")))
		return allFields
	}

	s = truncate(string(jsonBytes), maxMessageLogSizeBytes)
	allFields = append(allFields, log.String("initialRequestJSON", s))

	return allFields
}

// truncate shortens the string be to at most maxBytes bytes, appending a message indicating that the string was truncated if necessary.
//
// If maxBytes is negative, then the string is not truncated.
func truncate(s string, maxBytes int) string {
	if maxBytes < 0 {
		return s
	}

	bytesToTruncate := len(s) - maxBytes
	if bytesToTruncate > 0 {
		s = s[:maxBytes]
		s = fmt.Sprintf("%s...(truncated %d bytes)", s, bytesToTruncate)
	}

	return s
}

func isNonUTF8StringError(s *status.Status) bool {
	if s.Code() != codes.Internal {
		return false
	}

	return strings.Contains(s.Message(), "string field contains invalid UTF-8")
}
