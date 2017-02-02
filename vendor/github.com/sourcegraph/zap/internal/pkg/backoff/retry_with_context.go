package backoff

import (
	"context"
	"time"
)

// RetryNotifyWithContext calls notify function with the error and
// wait duration for each failed attempt before sleep. It will return
// early from a sleep when ctx's Done channel is closed, and it will
// also not call the operation if the context is already canceled.
func RetryNotifyWithContext(ctx context.Context, operation ContextOperation, b BackOff, notify Notify) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	var err error
	var next time.Duration

	b.Reset()
	for {
		if err = operation(ctx); err == nil {
			return nil
		} else if err == context.Canceled || err == context.DeadlineExceeded {
			return err
		}

		if next = b.NextBackOff(); next == Stop {
			return err
		}

		if notify != nil {
			notify(err, next)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(next):
		}
	}
}
