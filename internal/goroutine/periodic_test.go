package goroutine

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestPeriodicGoroutine(t *testing.T) {
	clock := glock.NewMockClock()
	handler := NewMockHandler()
	called := make(chan struct{}, 1)

	handler.HandleFunc.SetDefaultHook(func(ctx context.Context) error {
		called <- struct{}{}
		return nil
	})

	goroutine := newPeriodicGoroutine(context.Background(), t.Name(), "", func() time.Duration { return time.Second }, handler, nil, clock)
	go goroutine.Start()
	clock.BlockingAdvance(time.Second)
	<-called
	clock.BlockingAdvance(time.Second)
	<-called
	clock.BlockingAdvance(time.Second)
	<-called
	goroutine.Stop()

	if calls := len(handler.HandleFunc.History()); calls != 4 {
		t.Errorf("unexpected number of handler invocations. want=%d have=%d", 4, calls)
	}
}

func TestPeriodicGoroutineWithDynamicInterval(t *testing.T) {
	clock := glock.NewMockClock()
	handler := NewMockHandler()
	called := make(chan struct{}, 1)

	handler.HandleFunc.SetDefaultHook(func(ctx context.Context) error {
		called <- struct{}{}
		return nil
	})

	seconds := 1

	// intervals: 1 -> 2 -> 3 ...
	getInterval := func() time.Duration {
		duration := time.Duration(seconds) * time.Second
		seconds += 1
		return duration
	}

	goroutine := newPeriodicGoroutine(context.Background(), t.Name(), "", getInterval, handler, nil, clock)
	go goroutine.Start()
	clock.BlockingAdvance(time.Second)
	<-called
	clock.BlockingAdvance(2 * time.Second)
	<-called
	clock.BlockingAdvance(3 * time.Second)
	<-called
	goroutine.Stop()

	if calls := len(handler.HandleFunc.History()); calls != 4 {
		t.Errorf("unexpected number of handler invocations. want=%d have=%d", 4, calls)
	}
}

func TestPeriodicGoroutineError(t *testing.T) {
	clock := glock.NewMockClock()
	handler := NewMockHandlerWithErrorHandler()

	calls := 0
	called := make(chan struct{}, 1)
	handler.HandleFunc.SetDefaultHook(func(ctx context.Context) (err error) {
		if calls == 0 {
			err = errors.New("oops")
		}

		calls++
		called <- struct{}{}
		return err
	})

	goroutine := newPeriodicGoroutine(context.Background(), t.Name(), "", func() time.Duration { return time.Second }, handler, nil, clock)
	go goroutine.Start()
	clock.BlockingAdvance(time.Second)
	<-called
	clock.BlockingAdvance(time.Second)
	<-called
	clock.BlockingAdvance(time.Second)
	<-called
	goroutine.Stop()

	if calls := len(handler.HandleFunc.History()); calls != 4 {
		t.Errorf("unexpected number of handler invocations. want=%d have=%d", 4, calls)
	}

	if calls := len(handler.HandleErrorFunc.History()); calls != 1 {
		t.Errorf("unexpected number of error handler invocations. want=%d have=%d", 1, calls)
	}
}

func TestPeriodicGoroutineContextError(t *testing.T) {
	clock := glock.NewMockClock()
	handler := NewMockHandlerWithErrorHandler()

	called := make(chan struct{}, 1)
	handler.HandleFunc.SetDefaultHook(func(ctx context.Context) error {
		called <- struct{}{}
		<-ctx.Done()
		return ctx.Err()
	})

	goroutine := newPeriodicGoroutine(context.Background(), t.Name(), "", func() time.Duration { return time.Second }, handler, nil, clock)
	go goroutine.Start()
	<-called
	goroutine.Stop()

	if calls := len(handler.HandleFunc.History()); calls != 1 {
		t.Errorf("unexpected number of handler invocations. want=%d have=%d", 1, calls)
	}

	if calls := len(handler.HandleErrorFunc.History()); calls != 0 {
		t.Errorf("unexpected number of error handler invocations. want=%d have=%d", 0, calls)
	}
}

func TestPeriodicGoroutineFinalizer(t *testing.T) {
	clock := glock.NewMockClock()
	handler := NewMockHandlerWithFinalizer()

	called := make(chan struct{}, 1)
	handler.HandleFunc.SetDefaultHook(func(ctx context.Context) error {
		called <- struct{}{}
		return nil
	})

	goroutine := newPeriodicGoroutine(context.Background(), t.Name(), "", func() time.Duration { return time.Second }, handler, nil, clock)
	go goroutine.Start()
	clock.BlockingAdvance(time.Second)
	<-called
	clock.BlockingAdvance(time.Second)
	<-called
	clock.BlockingAdvance(time.Second)
	<-called
	goroutine.Stop()

	if calls := len(handler.HandleFunc.History()); calls != 4 {
		t.Errorf("unexpected number of handler invocations. want=%d have=%d", 4, calls)
	}

	if calls := len(handler.OnShutdownFunc.History()); calls != 1 {
		t.Errorf("unexpected number of finalizer invocations. want=%d have=%d", 1, calls)
	}
}

type MockHandlerWithErrorHandler struct {
	*MockHandler
	*MockErrorHandler
}

func NewMockHandlerWithErrorHandler() *MockHandlerWithErrorHandler {
	return &MockHandlerWithErrorHandler{
		MockHandler:      NewMockHandler(),
		MockErrorHandler: NewMockErrorHandler(),
	}
}

type MockHandlerWithFinalizer struct {
	*MockHandler
	*MockFinalizer
}

func NewMockHandlerWithFinalizer() *MockHandlerWithFinalizer {
	return &MockHandlerWithFinalizer{
		MockHandler:   NewMockHandler(),
		MockFinalizer: NewMockFinalizer(),
	}
}
