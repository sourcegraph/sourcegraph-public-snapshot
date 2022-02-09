package gqltestutil

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrContinueRetry = errors.New("continue Retry")

// Retry retries the given function until the timeout is reached. The function should
// return ErrContinueRetry to indicate another retry.
func Retry(timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return errors.Errorf("Retry timed out in %s", timeout)
			}
			return ctx.Err()
		default:
			err := fn()
			if err != ErrContinueRetry {
				return err
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}
