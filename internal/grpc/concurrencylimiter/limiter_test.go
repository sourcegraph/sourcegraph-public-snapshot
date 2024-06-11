package concurrencylimiter

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"

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
