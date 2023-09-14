package contextconv

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestConversionUnaryInterceptor(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	deadlineExceededCtx, cancel := context.WithDeadline(context.Background(), time.Now())
	t.Cleanup(cancel) // only cancel after all subtests have run, we want the context to be deadline exceeded in the relevant subtest

	testCases := []struct {
		name        string
		ctx         context.Context
		handlerErr  error
		expectedErr error
	}{
		{
			name:        "handler success",
			ctx:         context.Background(),
			handlerErr:  nil,
			expectedErr: nil,
		},
		{
			name:        "status error",
			ctx:         context.Background(),
			handlerErr:  status.Error(codes.Internal, "internal error"),
			expectedErr: status.Error(codes.Internal, "internal error"),
		},
		{
			name:        "context cancellation error",
			ctx:         cancelledCtx,
			handlerErr:  cancelledCtx.Err(),
			expectedErr: status.Error(codes.Canceled, "context canceled"),
		},
		{
			name:        "context deadline error",
			ctx:         deadlineExceededCtx,
			handlerErr:  deadlineExceededCtx.Err(),
			expectedErr: status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
		},
		{
			name:        "unknown error",
			ctx:         context.Background(),
			handlerErr:  errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sentinelResponse := struct{}{}

			called := false
			handler := func(ctx context.Context, request any) (response any, err error) {
				called = true
				return sentinelResponse, tc.handlerErr
			}

			request := struct{}{}
			info := grpc.UnaryServerInfo{}

			actualResponse, err := UnaryServerInterceptor(tc.ctx, request, &info, handler)
			if !called {
				t.Fatal("handler was not called")
			}

			if actualResponse != sentinelResponse {
				t.Fatalf("unexpected response: %+v", actualResponse)
			}

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			actualErrString := fmt.Sprintf("%s", err)

			if diff := cmp.Diff(expectedErrString, actualErrString); diff != "" {
				t.Errorf("unexpected error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStreamServerInterceptor(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	deadlineExceededCtx, cancel := context.WithDeadline(context.Background(), time.Now())
	t.Cleanup(cancel) // only cancel after all subtests have run, we want the context to be deadline exceeded in the relevant subtest

	testCases := []struct {
		name        string
		ctx         context.Context
		handlerErr  error
		expectedErr error
	}{
		{
			name:        "handler success",
			ctx:         context.Background(),
			handlerErr:  nil,
			expectedErr: nil,
		},
		{
			name:        "status error",
			ctx:         context.Background(),
			handlerErr:  status.Error(codes.Internal, "internal error"),
			expectedErr: status.Error(codes.Internal, "internal error"),
		},
		{
			name:        "context cancellation error",
			ctx:         cancelledCtx,
			handlerErr:  cancelledCtx.Err(),
			expectedErr: status.Error(codes.Canceled, "context canceled"),
		},
		{
			name:        "context deadline error",
			ctx:         deadlineExceededCtx,
			handlerErr:  deadlineExceededCtx.Err(),
			expectedErr: status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
		},
		{
			name:        "unknown error",
			ctx:         context.Background(),
			handlerErr:  errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			called := false

			handler := func(_ any, _ grpc.ServerStream) error {
				called = true
				return tc.handlerErr
			}

			srv := struct{}{}
			info := grpc.StreamServerInfo{}

			err := StreamServerInterceptor(srv, &mockServerStream{ctx: tc.ctx}, &info, handler)
			if !called {
				t.Fatal("handler was not called")
			}

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			actualErrString := fmt.Sprintf("%s", err)
			if diff := cmp.Diff(expectedErrString, actualErrString); diff != "" {
				t.Errorf("unexpected error (-want +got):\n%s", diff)
			}
		})
	}
}

// mockServerStream is a fake grpc.ServerStream that returns the provided context
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func TestUnaryClientInterceptor(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	deadlineExceededCtx, cancel := context.WithDeadline(context.Background(), time.Now())
	t.Cleanup(cancel)

	testCases := []struct {
		name        string
		ctx         context.Context
		invokerErr  error
		expectedErr error
	}{
		{
			name:        "invoker success",
			ctx:         context.Background(),
			invokerErr:  nil,
			expectedErr: nil,
		},
		{
			name:        "status error",
			ctx:         context.Background(),
			invokerErr:  status.Error(codes.Internal, "internal error"),
			expectedErr: status.Error(codes.Internal, "internal error"),
		},
		{
			name:        "context cancellation error",
			ctx:         cancelledCtx,
			invokerErr:  status.Error(codes.Canceled, "context canceled"),
			expectedErr: context.Canceled,
		},
		{
			name:        "context deadline error",
			ctx:         deadlineExceededCtx,
			invokerErr:  status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
			expectedErr: context.DeadlineExceeded,
		},
		{
			name:        "unknown error",
			ctx:         context.Background(),
			invokerErr:  errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
				called = true
				return tc.invokerErr
			}

			method := "Test"
			req := struct{}{}
			reply := struct{}{}
			cc := &grpc.ClientConn{}

			err := UnaryClientInterceptor(tc.ctx, method, req, &reply, cc, invoker)
			if !called {
				t.Fatal("invoker was not called")
			}

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			actualErrString := fmt.Sprintf("%s", err)
			if diff := cmp.Diff(expectedErrString, actualErrString); diff != "" {
				t.Errorf("unexpected error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStreamClientInterceptor(t *testing.T) {
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	deadlineExceededCtx, cancel := context.WithDeadline(context.Background(), time.Now())
	t.Cleanup(cancel)

	testCases := []struct {
		name        string
		ctx         context.Context
		streamerErr error
		expectedErr error
	}{
		{
			name:        "streamer success",
			ctx:         context.Background(),
			streamerErr: nil,
			expectedErr: nil,
		},
		{
			name:        "status error",
			ctx:         context.Background(),
			streamerErr: status.Error(codes.Internal, "internal error"),
			expectedErr: status.Error(codes.Internal, "internal error"),
		},
		{
			name:        "context cancellation error",
			ctx:         cancelledCtx,
			streamerErr: status.Error(codes.Canceled, "context canceled"),
			expectedErr: context.Canceled,
		},
		{
			name:        "context deadline error",
			ctx:         deadlineExceededCtx,
			streamerErr: status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
			expectedErr: context.DeadlineExceeded,
		},
		{
			name:        "unknown error",
			ctx:         context.Background(),
			streamerErr: errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			streamer := func(_ context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
				called = true
				if tc.streamerErr != nil {
					return nil, tc.streamerErr
				}
				return &mockClientStream{}, nil
			}

			desc := &grpc.StreamDesc{}
			cc := &grpc.ClientConn{}
			method := "Test"

			_, clientErr := StreamClientInterceptor(tc.ctx, desc, cc, method, streamer)
			if !called {
				t.Fatal("streamer was not called")
			}

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			actualErrString := fmt.Sprintf("%s", clientErr)
			if diff := cmp.Diff(expectedErrString, actualErrString); diff != "" {
				t.Fatalf("unexpected error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestConvertingClientStream(t *testing.T) {
	testCases := []struct {
		name        string
		streamErr   error
		expectedErr error
	}{
		{
			name:        "status error",
			streamErr:   status.Error(codes.Internal, "internal error"),
			expectedErr: status.Error(codes.Internal, "internal error"),
		},
		{
			name:        "context cancellation error",
			streamErr:   status.Error(codes.Canceled, "context canceled"),
			expectedErr: context.Canceled,
		},
		{
			name:        "context deadline error",
			streamErr:   status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
			expectedErr: context.DeadlineExceeded,
		},
		{
			name:        "unknown error",
			streamErr:   errors.New("unknown error"),
			expectedErr: errors.New("unknown error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStream := &mockClientStream{
				err: tc.streamErr,
			}
			wrappedStream := &convertingClientStream{
				ClientStream: mockStream,
			}

			err := wrappedStream.RecvMsg(nil)

			expectedErrString := fmt.Sprintf("%s", tc.expectedErr)
			actualErrString := fmt.Sprintf("%s", err)

			if diff := cmp.Diff(expectedErrString, actualErrString); diff != "" {
				t.Errorf("unexpected error (-want +got):\n%s", diff)
			}
		})
	}
}

// mockClientStream is a fake grpc.ClientStream
type mockClientStream struct {
	grpc.ClientStream
	err error
}

func (m *mockClientStream) RecvMsg(x interface{}) error {
	return m.err
}
