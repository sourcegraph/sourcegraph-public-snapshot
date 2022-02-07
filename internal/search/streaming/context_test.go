package streaming

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCancelWithReason(t *testing.T) {
	parentCtx, cancel := context.WithCancel(context.Background())
	mutCtx := WithMutableValue(parentCtx)
	childCtx, cancelChildCtx := IgnoreContextCancellation(mutCtx, CanceledLimitHit)
	defer cancelChildCtx()

	// Cancel with reason.
	mutCtx.Set(CanceledLimitHit, true)
	cancel()

	select {
	case <-parentCtx.Done():
	case <-time.After(5 * time.Second):
		t.Fatalf("parent context should have been canceled")
	}

	select {
	case <-childCtx.Done():
		t.Fatalf("child context should NOT have been canceled")
	case <-time.After(10 * time.Millisecond):
	}

	if childCtx.Err() != nil {
		t.Fatalf("child context should return nil error")
	}
}

func TestCancelWithoutReason(t *testing.T) {
	parentCtx, cancel := context.WithCancel(context.Background())
	mutCtx := WithMutableValue(parentCtx)
	childCtx, cancelChildCtx := IgnoreContextCancellation(mutCtx, CanceledLimitHit)
	defer cancelChildCtx()
	// Check propagation to children.
	type tmpType string
	grandchildCtx := context.WithValue(childCtx, tmpType("foo"), "bar")

	// We cancel without giving a reason and expect childCtx to be canceled.
	cancel()

	select {
	case <-childCtx.Done():
	case <-time.After(5 * time.Second):
		t.Fatalf("child context should have been canceled")
	}

	select {
	case <-grandchildCtx.Done():
	case <-time.After(5 * time.Second):
		t.Fatalf("child context should have been canceled")
	}

	if !errors.Is(childCtx.Err(), context.Canceled) {
		t.Fatalf("got %v, want %v", childCtx.Err(), context.Canceled)
	}

	if !errors.Is(grandchildCtx.Err(), context.Canceled) {
		t.Fatalf("got %v, want %v", grandchildCtx.Err(), context.Canceled)
	}
}

func TestDeadlineExceeded(t *testing.T) {
	parentCtx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()
	mutCtx := WithMutableValue(parentCtx)
	childCtx, cleanup := IgnoreContextCancellation(mutCtx, CanceledLimitHit)
	defer cleanup()

	select {
	case <-childCtx.Done():
	case <-time.After(5 * time.Second):
		t.Fatalf("child context should have been canceled")
	}

	if !errors.Is(childCtx.Err(), context.DeadlineExceeded) {
		t.Fatalf("got %v, want %v", childCtx.Err(), context.DeadlineExceeded)
	}
}

func TestCancelChildContext(t *testing.T) {
	parentCtx := context.Background()
	mutCtx := WithMutableValue(parentCtx)
	childCtx, cancelChildCtx := IgnoreContextCancellation(mutCtx, CanceledLimitHit)
	cancelChildCtx()

	select {
	case <-parentCtx.Done():
		t.Fatalf("parent context should not have been canceled")
	case <-time.After(10 * time.Millisecond):
	}

	select {
	case <-childCtx.Done():
	case <-time.After(5 * time.Second):
		t.Fatalf("child context should have been canceled")
	}

	if !errors.Is(childCtx.Err(), context.Canceled) {
		t.Fatalf("got %s, want %s\n", childCtx.Err(), context.Canceled)
	}
}
