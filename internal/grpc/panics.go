package grpc

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/sourcegraph/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newPanicErr(val any) error {
	return status.Errorf(codes.Internal, "panic during execution: %v", val)
}

func NewStreamPanicCatcher(logger log.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if val := recover(); val != nil {
				err = newPanicErr(val)
				logger.Error(
					fmt.Sprintf("caught panic: %s", string(debug.Stack())),
					log.String("method", info.FullMethod),
					log.Error(err),
				)
			}
		}()

		return handler(srv, ss)
	}
}

func NewUnaryPanicCatcher(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if val := recover(); val != nil {
				err = newPanicErr(val)
				logger.Error(
					fmt.Sprintf("caught panic: %s", string(debug.Stack())),
					log.String("method", info.FullMethod),
					log.Error(err),
				)
			}
		}()

		return handler(ctx, req)
	}
}
