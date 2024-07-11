package goroutine

import (
	"context"
	"fmt"
	"time"
)

type exampleRoutine struct {
	done chan struct{}
}

func (m *exampleRoutine) Name() string {
	return "exampleRoutine"
}

func (m *exampleRoutine) Start() {
	for {
		select {
		case <-m.done:
			fmt.Println("done!")
			return
		default:
		}

		fmt.Println("Hello there!")
		time.Sleep(200 * time.Millisecond)
	}
}

func (m *exampleRoutine) Stop(context.Context) error {
	m.done <- struct{}{}
	return nil
}

func ExampleBackgroundRoutine() {
	r := &exampleRoutine{
		done: make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		err := MonitorBackgroundRoutines(ctx, r)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()
	time.Sleep(200 * time.Millisecond)
}

func ExamplePeriodicGoroutine() {
	h := HandlerFunc(func(ctx context.Context) error {
		fmt.Println("Hello from the background!")
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())

	r := NewPeriodicGoroutine(
		ctx,
		h,
		WithName("example.background"),
		WithDescription("example background routine"),
		WithInterval(200*time.Millisecond),
	)

	go func() {
		err := MonitorBackgroundRoutines(ctx, r)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	cancel()
	time.Sleep(200 * time.Millisecond)
}
