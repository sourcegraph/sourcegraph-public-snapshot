package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/efritz/glock"
)

func TestRetry(t *testing.T) {
	errs := []error{
		fmt.Errorf("err1: read: connection reset by peer"),
		fmt.Errorf("err2: read: connection reset by peer"),
		fmt.Errorf("err3: read: connection reset by peer"),
	}

	calls := 0
	retryable := func(ctx context.Context) error {
		calls++

		if len(errs) == 0 {
			return nil
		}

		err := errs[0]
		errs = errs[1:]
		return err
	}

	clock := glock.NewMockClock()

	go func() {
		clock.BlockingAdvance(minBackoff)
		clock.BlockingAdvance(time.Duration(float64(minBackoff) * backoffIncreaseRatio))
		clock.BlockingAdvance(time.Duration(float64(minBackoff) * backoffIncreaseRatio * backoffIncreaseRatio))
	}()

	if err := retry(context.Background(), clock, retryable); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if calls != 4 {
		t.Errorf("unexpected number of calls. want=%d have=%d", 4, calls)
	}
}

func TestRetryNonRetryableError(t *testing.T) {
	expectedErr := fmt.Errorf("oops")
	retryable := func(ctx context.Context) error {
		return expectedErr
	}

	clock := glock.NewMockClock()

	if err := retry(context.Background(), clock, retryable); err != expectedErr {
		t.Fatalf("unexpected error. want=%q have=%q", expectedErr, err)
	}
}

func TestRetryMaxAttempts(t *testing.T) {
	expectedErr := errors.New("read: connection reset by peer")
	retryable := func(ctx context.Context) error {
		return expectedErr
	}

	if err := retry(context.Background(), advancingClock(), retryable); err != expectedErr {
		t.Fatalf("unexpected error. want=%q have=%q", expectedErr, err)
	}
}

func TestRetryContextCanceled(t *testing.T) {
	expectedErr := errors.New("read: connection reset by peer")
	retryable := func(ctx context.Context) error {
		return expectedErr
	}

	clock := glock.NewMockClock()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := retry(ctx, clock, retryable); err != expectedErr {
		t.Fatalf("unexpected error. want=%q have=%q", expectedErr, err)
	}
}

// advancingClock returns a mock clock that advances by maxBackoff in a loop.
func advancingClock() glock.Clock {
	clock := glock.NewMockClock()

	go func() {
		for {
			clock.BlockingAdvance(maxBackoff)
		}
	}()

	return clock
}
