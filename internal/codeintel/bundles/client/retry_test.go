package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/efritz/glock"
)

func TestRetry(t *testing.T) {
	calls := 0
	errs := []error{fmt.Errorf("err1"), fmt.Errorf("err2"), fmt.Errorf("err3")}
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

	if err := retryWithClock(context.Background(), clock, retryable); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if calls != 4 {
		t.Errorf("unexpected number of calls. want=%d have=%d", 4, calls)
	}
}

func TestRetryFailure(t *testing.T) {
	expectedErr := fmt.Errorf("oops")
	retryable := func(ctx context.Context) error {
		return expectedErr
	}

	clock := glock.NewMockClock()

	go func() {
		for {
			clock.BlockingAdvance(maxBackoff)
		}
	}()

	if err := retryWithClock(context.Background(), clock, retryable); err != expectedErr {
		t.Fatalf("unexpected error. want=%q have=%q", expectedErr, err)
	}
}

func TestRetryContextCanceled(t *testing.T) {
	expectedErr := fmt.Errorf("oops")
	retryable := func(ctx context.Context) error {
		return expectedErr
	}

	clock := glock.NewMockClock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := retryWithClock(ctx, clock, retryable); err != expectedErr {
		t.Fatalf("unexpected error. want=%q have=%q", expectedErr, err)
	}
}
