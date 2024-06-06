package concurrencylimiter

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/limiter"
)

func TestUnaryClientInterceptor(t *testing.T) {
	t.Run("acquire and release limiter", func(t *testing.T) {
		limiter := &mockLimiter{}
		in := UnaryClientInterceptor(limiter)

		invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			return nil
		}

		_ = in(context.Background(), "Test", struct{}{}, struct{}{}, &grpc.ClientConn{}, invoker)

		if limiter.acquireCount != 1 {
			t.Errorf("expected acquire count to be 1, got %d", limiter.acquireCount)
		}
		if limiter.releaseCount != 1 {
			t.Errorf("expected release count to be 1, got %d", limiter.releaseCount)
		}
	})

	t.Run("invoker error propagated", func(t *testing.T) {
		limiter := &mockLimiter{}
		in := UnaryClientInterceptor(limiter)

		expectedErr := errors.New("invoker error")
		invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			return expectedErr
		}

		err := in(context.Background(), "Test", struct{}{}, struct{}{}, &grpc.ClientConn{}, invoker)

		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if limiter.acquireCount != 1 {
			t.Errorf("expected acquire count to be 1, got %d", limiter.acquireCount)
		}
		if limiter.releaseCount != 1 {
			t.Errorf("expected release count to be 1, got %d", limiter.releaseCount)
		}
	})

	t.Run("maximum concurrency honored", func(t *testing.T) {
		limiter := limiter.New(1)
		in := UnaryClientInterceptor(limiter)

		var wg sync.WaitGroup
		count := 3000
		wg.Add(count)

		var concurrency atomic.Int32
		var notHonored atomic.Bool

		invoker := func(_ context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			defer concurrency.Add(-1)

			if !concurrency.CompareAndSwap(0, 1) {
				notHonored.Store(true)
			}

			return nil
		}

		for range count {
			go func() {
				_ = in(context.Background(), "Test", struct{}{}, struct{}{}, &grpc.ClientConn{}, invoker)
				wg.Done()
			}()
		}

		wg.Wait()

		if notHonored.Load() {
			t.Fatal("concurrency limit not honored")
		}
	})
}

func TestStreamClientInterceptor(t *testing.T) {
	t.Run("acquire and release limiter", func(t *testing.T) {
		limiter := &mockLimiter{}
		in := StreamClientInterceptor(limiter)

		streamer := func(_ context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
			return &mockClientStream{err: io.EOF}, nil
		}

		cc, _ := in(context.Background(), &grpc.StreamDesc{}, &grpc.ClientConn{}, "Test", streamer)

		require.NoError(t, cc.SendMsg(nil))
		require.NoError(t, cc.CloseSend())
		require.Equal(t, io.EOF, cc.RecvMsg(nil))

		if limiter.acquireCount != 1 {
			t.Errorf("expected acquire count to be 1, got %d", limiter.acquireCount)
		}
		if limiter.releaseCount != 1 {
			t.Errorf("expected release count to be 1, got %d", limiter.releaseCount)
		}
	})

	t.Run("streamer error propagated", func(t *testing.T) {
		limiter := &mockLimiter{}
		in := StreamClientInterceptor(limiter)

		expectedErr := errors.New("streamer error")
		streamer := func(_ context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
			return nil, expectedErr
		}

		_, err := in(context.Background(), &grpc.StreamDesc{}, &grpc.ClientConn{}, "Test", streamer)

		if err != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
		if limiter.acquireCount != 1 {
			t.Errorf("expected acquire count to be 1, got %d", limiter.acquireCount)
		}
		if limiter.releaseCount != 1 {
			t.Errorf("expected release count to be 1, got %d", limiter.releaseCount)
		}

		limiter = &mockLimiter{}
		in = StreamClientInterceptor(limiter)
		streamer = func(_ context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
			return &mockClientStream{err: expectedErr}, nil
		}
		cc, err := in(context.Background(), &grpc.StreamDesc{}, &grpc.ClientConn{}, "Test", streamer)
		require.NoError(t, err)

		require.NoError(t, cc.SendMsg(nil))
		require.NoError(t, cc.CloseSend())
		require.Equal(t, expectedErr, cc.RecvMsg(nil))
		if limiter.acquireCount != 1 {
			t.Errorf("expected acquire count to be 1, got %d", limiter.acquireCount)
		}
		if limiter.releaseCount != 1 {
			t.Errorf("expected release count to be 1, got %d", limiter.releaseCount)
		}
	})

	t.Run("maximum concurrency honored", func(t *testing.T) {
		limiter := limiter.New(1)
		in := StreamClientInterceptor(limiter)

		var wg sync.WaitGroup
		count := 3000
		wg.Add(count)

		var concurrency atomic.Int32
		var notHonored atomic.Bool

		streamer := func(_ context.Context, _ *grpc.StreamDesc, _ *grpc.ClientConn, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
			if !concurrency.CompareAndSwap(0, 1) {
				notHonored.Store(true)
			}

			return &limitedClientStream{
				release: func() {
					concurrency.Add(-1)
				},
				ClientStream: &mockClientStream{err: io.EOF},
			}, nil
		}

		for range count {
			go func() {
				cc, _ := in(context.Background(), &grpc.StreamDesc{}, &grpc.ClientConn{}, "Test", streamer)
				require.NoError(t, cc.SendMsg(nil))
				require.NoError(t, cc.CloseSend())
				require.Equal(t, io.EOF, cc.RecvMsg(nil))
				wg.Done()
			}()
		}

		wg.Wait()

		if notHonored.Load() {
			t.Fatal("concurrency limit not honored")
		}
	})
}

type mockLimiter struct {
	acquireCount int
	releaseCount int
}

func (m *mockLimiter) Acquire() {
	m.acquireCount++
}

func (m *mockLimiter) Release() {
	m.releaseCount++
}

type mockClientStream struct {
	grpc.ClientStream
	err error
}

func (m *mockClientStream) CloseSend() error {
	return nil
}

func (m *mockClientStream) SendMsg(x interface{}) error {
	return nil
}

func (m *mockClientStream) RecvMsg(x interface{}) error {
	return m.err
}
