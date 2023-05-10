package internalerrs

import (
	"context"
	"io"
	"strings"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	logScope       = "gRPC.internal.error.reporter"
	logDescription = "logs gRPC errors that appear to come from the go-grpc implementation"
)

func LoggingUnaryClientInterceptor(logger log.Logger) grpc.UnaryClientInterceptor {
	logger = logger.Scoped(logScope, logDescription)

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		serviceName, methodName := splitMethodName(method)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			doLog(logger, serviceName, methodName, err)
		}

		return err
	}
}

func LoggingStreamClientInterceptor(logger log.Logger) grpc.StreamClientInterceptor {
	logger = logger.Scoped(logScope, logDescription)

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		serviceName, methodName := splitMethodName(method)

		stream, err := streamer(ctx, desc, cc, method, opts...)
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
			if err != nil && err != io.EOF {
				doLog(logger, serviceName, methodName, err)
			}
		},
	}
}

type callBackClientStream struct {
	grpc.ClientStream

	postMessageSend    func(error)
	postMessageReceive func(error)
}

func (c *callBackClientStream) SendMsg(m interface{}) error {
	err := c.ClientStream.SendMsg(m)
	c.postMessageSend(err)

	return err
}

func (c *callBackClientStream) RecvMsg(m interface{}) error {
	err := c.ClientStream.RecvMsg(m)
	c.postMessageReceive(err)

	return err
}

var _ grpc.ClientStream = &callBackClientStream{}

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

func probablyInternalGRPCError(s *status.Status) bool {
	return s.Code() != codes.OK && strings.HasPrefix(s.Message(), "grpc:")
}

// splitMethodName splits a full gRPC method name in to its components (service, method)
//
// Copied from github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/reporter.go
func splitMethodName(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/") // remove leading slash
	if i := strings.Index(fullMethod, "/"); i >= 0 {
		return fullMethod[:i], fullMethod[i+1:]
	}
	return "unknown", "unknown"
}
