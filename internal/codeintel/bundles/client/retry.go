package client

import (
	"context"
	"time"

	"github.com/efritz/glock"
)

// TODO - document
type Retryable func(ctx context.Context) error

// TODO - document
func retry(ctx context.Context, f Retryable) error {
	return retryWithClock(ctx, glock.NewRealClock(), f)
}

const (
	minBackoff           = time.Second * 1
	maxBackoff           = time.Second * 5
	backoffIncreaseRatio = 1.5
	maxAttempts          = 5
)

func retryWithClock(ctx context.Context, clock glock.Clock, f Retryable) (err error) {
	backoff := minBackoff

	for attempts := maxAttempts; attempts > 0; attempts-- {
		// TODO - should match err against isConnectionError?
		if err = f(ctx); err == nil {
			return nil
		}

		select {
		case <-clock.After(backoff):
			if backoff = time.Duration(float64(backoff) * backoffIncreaseRatio); backoff > maxBackoff {
				backoff = maxBackoff
			}

		case <-ctx.Done():
			break
		}
	}

	return err
}
