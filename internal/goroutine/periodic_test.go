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

	goroutine := NewPeriodicGoroutine(
		context.Background(),
		handler,
		WithName(t.Name()),
		WithInterval(time.Second),
		withClock(clock),
	)
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

func TestPeriodicGoroutineReinvoke(t *testing.T) {
	clock := glock.NewMockClock()
	handler := NewMockHandler()
	called := make(chan struct{}, 1)

	handler.HandleFunc.SetDefaultHook(func(ctx context.Context) error {
		called <- struct{}{}
		return ErrReinvokeImmediately
	})

	witnessHandler := func() {
		for i := 0; i < maxConsecutiveReinvocations; i++ {
			<-called
		}
	}

	goroutine := NewPeriodicGoroutine(
		context.Background(),
		handler,
		WithName(t.Name()),
		WithInterval(time.Second),
		withClock(clock),
	)
	go goroutine.Start()
	witnessHandler()
	clock.BlockingAdvance(time.Second)
	witnessHandler()
	clock.BlockingAdvance(time.Second)
	witnessHandler()
	clock.BlockingAdvance(time.Second)
	witnessHandler()
	goroutine.Stop()

	if calls := len(handler.HandleFunc.History()); calls != 4*maxConsecutiveReinvocations {
		t.Errorf("unexpected number of handler invocations. want=%d have=%d", 4*maxConsecutiveReinvocations, calls)
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

	goroutine := NewPeriodicGoroutine(
		context.Background(),
		handler,
		WithName(t.Name()),
		WithIntervalFunc(getInterval),
		withClock(clock),
	)
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

func TestPeriodicGoroutineConcurrency(t *testing.T) {
	clock := glock.NewMockClock()
	handler := NewMockHandler()
	called := make(chan struct{}, 1)
	concurrency := 4

	handler.HandleFunc.SetDefaultHook(func(ctx context.Context) error {
		called <- struct{}{}
		return nil
	})

	goroutine := NewPeriodicGoroutine(
		context.Background(),
		handler,
		WithName(t.Name()),
		WithConcurrency(concurrency),
		withClock(clock),
	)
	go goroutine.Start()
	clock.BlockingAdvance(time.Second)
	for i := 0; i < concurrency; i++ {
		<-called
	}
	clock.BlockingAdvance(time.Second)
	for i := 0; i < concurrency; i++ {
		<-called
	}
	clock.BlockingAdvance(time.Second)
	for i := 0; i < concurrency; i++ {
		<-called
	}
	goroutine.Stop()

	if calls := len(handler.HandleFunc.History()); calls != 3*concurrency {
		t.Errorf("unexpected number of handler invocations. want=%d have=%d", 3*concurrency, calls)
	}
}

func TestPeriodicGoroutineWithDynamicConcurrency(t *testing.T) {
	clock := glock.NewMockClock()
	concurrencyClock := glock.NewMockClock()
	handler := NewMockHandler()
	called := make(chan struct{}, 1)

	handler.HandleFunc.SetDefaultHook(func(ctx context.Context) error {
		called <- struct{}{}
		return nil
	})

	concurrency := 0

	// concurrency: 1 -> 2 -> 3 ...
	getConcurrency := func() int {
		concurrency += 1
		return concurrency
	}

	goroutine := NewPeriodicGoroutine(
		context.Background(),
		handler,
		WithName(t.Name()),
		WithConcurrencyFunc(getConcurrency),
		withClock(clock),
		withConcurrencyClock(concurrencyClock),
	)
	go goroutine.Start()

	// Pool size = 1
	<-called

	// Pool size = 2
	concurrencyClock.BlockingAdvance(concurrencyRecheckInterval)
	<-called
	<-called
	clock.BlockingAdvance(time.Second)
	<-called
	<-called

	// Pool size 3
	concurrencyClock.BlockingAdvance(concurrencyRecheckInterval)
	<-called
	<-called
	<-called
	clock.BlockingAdvance(time.Second)
	<-called
	<-called
	<-called

	goroutine.Stop()

	if calls := len(handler.HandleFunc.History()); calls != 11 {
		t.Errorf("unexpected number of handler invocations. want=%d have=%d", 11, calls)
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

	goroutine := NewPeriodicGoroutine(
		context.Background(),
		handler,
		WithName(t.Name()),
		WithInterval(time.Second),
		withClock(clock),
	)
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

	goroutine := NewPeriodicGoroutine(
		context.Background(),
		handler,
		WithName(t.Name()),
		WithInterval(time.Second),
		withClock(clock),
	)
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

	goroutine := NewPeriodicGoroutine(
		context.Background(),
		handler,
		WithName(t.Name()),
		WithInterval(time.Second),
		withClock(clock),
	)
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
