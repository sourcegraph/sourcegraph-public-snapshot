package limiter

import (
	"context"
	"testing"
	"time"
)

func TestMutableLimiter(t *testing.T) {
	// cancels created by helpers
	var cancels []context.CancelFunc
	defer func() {
		for _, f := range cancels {
			f()
		}
	}()

	timeoutContext := func(d time.Duration) context.Context {
		ctx, cancel := context.WithTimeout(context.Background(), d)
		cancels = append(cancels, cancel)
		return ctx
	}

	l := NewMutable(2)

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
	_, _, err = l.Acquire(timeoutContext(250 * time.Millisecond))
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
	ctx3, cancel3, err := l.Acquire(timeoutContext(10 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer cancel3()

	// Adjust limit down, should cancel oldest job
	l.SetLimit(2)
	select {
	case <-ctx1.Done():
		// what we want
	case <-time.After(5 * time.Second):
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
