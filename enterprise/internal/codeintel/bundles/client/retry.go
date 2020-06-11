package client

import (
	"context"
	"net"
	"time"

	"github.com/efritz/glock"
)

const (
	minBackoff           = time.Second * 1
	maxBackoff           = time.Second * 5
	backoffIncreaseRatio = 1.5
	maxAttempts          = 20
)

// Retryable denotes an idempotent function that can be re-invoked on certain classes
// of errors. It is important that additional invocations of this function do not modify
// global state in a way that is harmful if done more than once, and does not destructively
// consume its input (e.g. a reader that cannot be seeked back to the beginning on error).
type Retryable func(ctx context.Context) error

// retry attempts to invoke the given retryable function until success. The last error
// returned from the retryable function will be returned after a maximum number of attempts,
// or if the given context has been canceled.
func retry(ctx context.Context, clock glock.Clock, f Retryable) (err error) {
	backoff := minBackoff

loop:
	for attempts := maxAttempts; attempts > 0; attempts-- {
		if err = f(ctx); err == nil || !isRetryableError(err) {
			return err
		}

		select {
		case <-clock.After(backoff):
			if backoff = time.Duration(float64(backoff) * backoffIncreaseRatio); backoff > maxBackoff {
				backoff = maxBackoff
			}

		case <-ctx.Done():
			break loop
		}
	}

	return err
}

// isRetryableError determines if the operation returning the given given non-nil error
// should be re-attempted. This is constrainted to a set of transient network errors that
// are problematic under heavy transfer load or due to deployment timings.
func isRetryableError(err error) bool {
	if _, ok := err.(net.Error); ok || isConnectionError(err) {
		return true
	}

	return false
}
