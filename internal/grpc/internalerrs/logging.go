package internalerrs

import (
	"context"
	"strings"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingUnaryClientInterceptor(logger log.Logger) grpc.UnaryClientInterceptor {
	logger = logger.Scoped("gRPC.internal.error.reporter", "")

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			s, ok := status.FromError(err)
			if !ok {
				return err
			}

			if !probablyInternalGRPCError(s) {
				return err
			}

			service, method := splitMethodName(method)
			logger.Error(s.Message(),
				log.String("grpc_service", service),
				log.String("grpc_method", method),
				log.String("grpc_code", s.Code().String()),
			)
		}

		if err != nil {

		}

		return err
	}
}

func LoggingStreamClientInterceptor(logger log.Logger) grpc.StreamClientInterceptor {
	logger = logger.Scoped("gRPC.internal.error.reporter", "logs ")

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		stream, err := streamer(ctx, desc, cc, method, opts...)
		stream = newLoggingStream(stream, logger, method)

		if err != nil {
			s, ok := status.FromError(err)
			if !ok {
				return stream, err
			}

			if !probablyInternalGRPCError(s) {
				return stream, err
			}

			service, method := splitMethodName(method)
			logger.Error(s.Message(),
				log.String("grpc_service", service),
				log.String("grpc_method", method),
				log.String("grpc_code", s.Code().String()))
		}

		return stream, err
	}
}

type loggingStream struct {
	grpc.ClientStream

	method  string
	service string

	logger log.Logger
}

func newLoggingStream(s grpc.ClientStream, logger log.Logger, fullMethod string) *loggingStream {
	service, method := splitMethodName(fullMethod)

	return &loggingStream{
		ClientStream: s,

		service: service,
		method:  method,

		logger: logger,
	}
}

func (l *loggingStream) RecvMsg(m any) error {
	err := l.ClientStream.RecvMsg(m)
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return err
		}

		if !probablyInternalGRPCError(s) {
			return err
		}

		l.logger.Error(s.Message(),
			log.String("grpc_service", l.service),
			log.String("grpc_method", l.method),
			log.String("grpc_code", s.Code().String()))
	}

	return err
}

func (l *loggingStream) SendMsg(m any) error {
	err := l.ClientStream.SendMsg(m)
	if err != nil {
		s, ok := status.FromError(err)
		if !ok {
			return err
		}

		if !probablyInternalGRPCError(s) {
			return err
		}

		l.logger.Error(s.Message(),
			log.String("grpc_service", l.service),
			log.String("grpc_method", l.method),
			log.String("grpc_code", s.Code().String()))
	}

	return err
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
