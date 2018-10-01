package mutablelimiter

import (
	"context"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	// cancels created by helpers
	var cancels []context.CancelFunc
	defer func() {
		for _, f := range cancels {
			f()
		}
	}()

	timeoutContext := func() context.Context {
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		cancels = append(cancels, cancel)
		return ctx
	}

	l := New(2)

	// Should not block
	ctx1, cancel1, err := l.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel1()
	ctx2, cancel2, err := l.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel2()

	// Should block, so use a context with a deadline
	_, _, err = l.Acquire(timeoutContext())
	if err != context.DeadlineExceeded {
		t.Fatal("expected acquire to fail")
	}

	l.SetLimit(3)

	// verify cap/len
	cap, len := l.GetLimit()
	if cap != 3 {
		t.Fatal("capacity not 3 as expected")
	}
	if len != 2 {
		t.Fatal("len not 2 as expected")
	}

	// Now should work. Still use context with a deadline to ensure acquire
	// wins over deadline
	ctx3, cancel3, err := l.Acquire(timeoutContext())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel3()

	// Adjust limit down, should cancel oldest job
	l.SetLimit(2)
	time.Sleep(100 * time.Millisecond) // give time to cancel
	if ctx1.Err() == nil {
		t.Fatal("expected first context to be canceled")
	}
	if ctx2.Err() != nil || ctx3.Err() != nil {
		t.Fatal("expected other contexts to still be running")
	}

	// Cancel 3rd job, should be able to then add another job
	cancel3()
	_, cancel4, err := l.Acquire(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer cancel4()
}
