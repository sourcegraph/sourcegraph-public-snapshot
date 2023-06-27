package contextconv

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor is a grpc.UnaryServerInterceptor that returns an appropriate status.Cancelled / status.DeadlineExceeded error
// if the handler call failed and the provided context has been cancelled or expired.
//
// The handler's error is propagated as-is if the context is still active or if the error is already one produced by the status package.
func UnaryServerInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response any, err error) {
	response, err = handler(ctx, req)
	if err == nil {
		return response, nil
	}

	if _, ok := status.FromError(err); ok {
		return response, err
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return response, status.FromContextError(ctxErr).Err()
	}

	return response, err
}

// StreamServerInterceptor is a grpc.StreamServerInterceptor that returns an appropriate status.Cancelled / status.DeadlineExceeded error
// if the handler call failed and the provided context has been cancelled or expired.
//
// The handler's error is propagated as-is if the context is still active or if the error is already one produced by the status package.
func StreamServerInterceptor(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, ss)
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	if ctxErr := ss.Context().Err(); ctxErr != nil {
		return status.FromContextError(ctxErr).Err()
	}

	return err
}

// UnaryClientInterceptor is a grpc.UnaryClientInterceptor that returns an appropriate context.DeadlineExceeded or context.Cancelled error
// if the call failed with a status.DeadlineExceeded or status.Cancelled error.
//
// The call's error is propagated as-is if the error is not status.DeadlineExceeded or status.Cancelled.
func UnaryClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err == nil {
		return nil
	}

	switch status.Code(err) {
	case codes.DeadlineExceeded:
		return context.DeadlineExceeded
	case codes.Canceled:
		return context.Canceled
	default:
		return err
	}
}

// StreamClientInterceptor is a grpc.StreamClientInterceptor that returns an appropriate context.DeadlineExceeded or context.Cancelled error
// if the call failed with a status.DeadlineExceeded or status.Cancelled error.
//
// The call's error is propagated as-is if the error is not status.DeadlineExceeded or status.Cancelled.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	clientStream, err := streamer(ctx, desc, cc, method, opts...)
	if err == nil {
		return &convertingClientStream{ClientStream: clientStream}, nil
	}

	switch status.Code(err) {
	case codes.DeadlineExceeded:
		return nil, context.DeadlineExceeded
	case codes.Canceled:
		return nil, context.Canceled
	default:
		return &convertingClientStream{ClientStream: clientStream}, err
	}
}

type convertingClientStream struct {
	grpc.ClientStream
}

func (c *convertingClientStream) RecvMsg(m any) error {
	err := c.ClientStream.RecvMsg(m)
	if err == nil {
		return nil
	}

	switch status.Code(err) {
	case codes.DeadlineExceeded:
		return context.DeadlineExceeded
	case codes.Canceled:
		return context.Canceled
	default:
		return err
	}
}

var (
	_ grpc.UnaryServerInterceptor  = UnaryServerInterceptor
	_ grpc.StreamServerInterceptor = StreamServerInterceptor

	_ grpc.UnaryClientInterceptor  = UnaryClientInterceptor
	_ grpc.StreamClientInterceptor = StreamClientInterceptor

	_ grpc.ClientStream = &convertingClientStream{}
)
