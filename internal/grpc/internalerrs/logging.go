package internalerrs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcutil"

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

	envLogMessagesEnabled                   = env.MustGetBool("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_PROTOBUF_MESSAGES_ENABLED", false, "Enables inclusion of raw protobuf messages in the gRPC internal error logs")
	envLogMessagesHandleMaxMessageSizeBytes = env.MustGetBytes("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_PROTOBUF_MESSAGES_HANDLING_MAX_MESSAGE_SIZE_BYTES", "100MB", "Maximum size of protobuf messages that can be included in gRPC internal error logs. The purpose of this is to avoid excessive allocations. 0 bytes mean no limit.")
	envLogMessagesMaxJSONSizeBytes          = env.MustGetBytes("SRC_GRPC_INTERNAL_ERROR_LOGGING_LOG_PROTOBUF_MESSAGES_JSON_TRUNCATION_SIZE_BYTES", "1KB", "Maximum size of the JSON representation of protobuf messages to log. JSON representations larger than this value will be truncated. 0 bytes disables truncation.")
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

	logger := l.Scoped(logScope)
	logger = logger.Scoped("unaryMethod")

	return func(ctx context.Context, fullMethod string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, fullMethod, req, reply, cc, opts...)
		if err != nil {
			serviceName, methodName := grpcutil.SplitMethodName(fullMethod)

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

	logger := l.Scoped(logScope)
	logger = logger.Scoped("streamingMethod")

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, fullMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		serviceName, methodName := grpcutil.SplitMethodName(fullMethod)

		stream, err := streamer(ctx, desc, cc, fullMethod, opts...)
		if err != nil {
			// Note: This is a bit hacky, we provide nil initial and payload messages here since the message isn't available
			// until after the stream is created.
			//
			// This is fine since the error is already available, and the non-utf8 string check is robust against nil messages.
			logger := logger.Scoped("postInit")
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

	logger := l.Scoped(logScope)
	logger = logger.Scoped("unaryMethod")

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		response, err := handler(ctx, req)
		if err != nil {
			serviceName, methodName := grpcutil.SplitMethodName(info.FullMethod)

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

	logger := l.Scoped(logScope)
	logger = logger.Scoped("streamingMethod")

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		serviceName, methodName := grpcutil.SplitMethodName(info.FullMethod)

		stream := newLoggingServerStream(ss, logger, serviceName, methodName)
		return handler(srv, stream)
	}
}

func newLoggingServerStream(s grpc.ServerStream, logger log.Logger, serviceName, methodName string) grpc.ServerStream {
	sendLogger := logger.Scoped("postMessageSend")
	receiveLogger := logger.Scoped("postMessageReceive")

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
	sendLogger := logger.Scoped("postMessageSend")
	receiveLogger := logger.Scoped("postMessageReceive")

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

	s, ok := massageIntoStatusErr(err)
	if !ok {
		// If the error isn't a grpc error, we don't know how to handle it.
		// Just return.
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

	// Log the initial request message
	if envLogMessagesEnabled {
		fs := messageJSONFields(initialRequest, "initialRequestJSON", envLogMessagesHandleMaxMessageSizeBytes, envLogMessagesMaxJSONSizeBytes)
		allFields = append(allFields, fs...)
	}

	if isNonUTF8StringError(s) {
		m, ok := payload.(proto.Message)
		if ok {
			allFields = append(allFields, nonUTF8StringLogFields(m)...)

			if envLogMessagesEnabled { // Log the latest message as well for non-utf8 errors
				fs := messageJSONFields(&m, "messageJSON", envLogMessagesHandleMaxMessageSizeBytes, envLogMessagesMaxJSONSizeBytes)
				allFields = append(allFields, fs...)
			}
		}
	}

	logger.Error(s.Message(), allFields...)
}

// messageJSONFields converts a protobuf message to a JSON string and returns it as a log field using the provided "key".
// The resulting JSON string is truncated to maxJSONSizeBytes.
//
// If the size of the original protobuf message exceeds maxMessageSizeBytes or any serialization errors are encountered, log fields
// describing the error are returned instead.
func messageJSONFields(m *proto.Message, key string, maxMessageSizeBytes, maxJSONSizeBytes uint64) []log.Field {
	if m == nil || *m == nil {
		return nil
	}

	if maxMessageSizeBytes > 0 {
		size := uint64(proto.Size(*m))
		if size > maxMessageSizeBytes {
			err := errors.Newf(
				"failed to marshal protobuf message (key: %q) to string: message too large (size %q, limit %q)",
				key,
				humanize.Bytes(size), humanize.Bytes(maxMessageSizeBytes),
			)

			return []log.Field{log.Error(err)}
		}
	}

	// Note: we can't use the protojson library here since it doesn't support messages with non-UTF8 strings.
	bs, err := json.Marshal(*m)
	if err != nil {
		err := errors.Wrapf(err, "failed to marshal protobuf message (key: %q) to string", key)
		return []log.Field{log.Error(err)}
	}

	s := truncate(string(bs), maxJSONSizeBytes)
	return []log.Field{log.String(key, s)}
}

// truncate shortens the string be to at most maxBytes bytes, appending a message indicating that the string was truncated if necessary.
//
// If maxBytes is 0, then the string is not truncated.
func truncate(s string, maxBytes uint64) string {
	if maxBytes <= 0 {
		return s
	}

	bytesToTruncate := len(s) - int(maxBytes)
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

// nonUTF8StringLogFields checks a protobuf message for fields that contain non-utf8 strings, and returns them as log fields.
func nonUTF8StringLogFields(m proto.Message) []log.Field {
	fs, err := findNonUTF8StringFields(m)
	if err != nil {
		err := errors.Wrapf(err, "failed to find non-UTF8 string fields in protobuf message")
		return []log.Field{log.Error(err)}

	}

	return []log.Field{log.Strings("nonUTF8StringFields", fs)}
}
