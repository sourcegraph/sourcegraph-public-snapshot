// package concurrencylimiter provides a concurrency limiter for use with grpc.
// The limiter is used to limit the number of concurrent calls to a grpc server.
package concurrencylimiter

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc"
)

// Limiter is a concurrency limiter. Acquire() blocks if the limit has been reached.
type Limiter interface {
	Acquire()
	Release()
}

// UnaryClientInterceptor returns a UnaryClientInterceptor that limits the number
// of concurrent calls with the given limiter.
func UnaryClientInterceptor(limiter Limiter) func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		limiter.Acquire()
		defer limiter.Release()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a StreamClientInterceptor that limits the number
// of concurrent calls with the given limiter.
func StreamClientInterceptor(limiter Limiter) func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		limiter.Acquire()

		var released int32
		release := func() {
			if atomic.CompareAndSwapInt32(&released, 0, 1) {
				limiter.Release()
			}
		}

		opts = append(opts, grpc.OnFinish(func(err error) {
			release()
		}))

		cs, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			release()
			return cs, err
		}

		return cs, nil
	}
}
