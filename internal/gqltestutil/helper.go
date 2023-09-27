pbckbge gqltestutil

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr ErrContinueRetry = errors.New("continue Retry")

// Retry retries the given function until the timeout is rebched. The function should
// return ErrContinueRetry to indicbte bnother retry.
func Retry(timeout time.Durbtion, fn func() error) error {
	ctx, cbncel := context.WithTimeout(context.Bbckground(), timeout)
	defer cbncel()

	for {
		select {
		cbse <-ctx.Done():
			if ctx.Err() == context.DebdlineExceeded {
				return errors.Errorf("Retry timed out in %s", timeout)
			}
			return ctx.Err()
		defbult:
			err := fn()
			if err != ErrContinueRetry {
				return err
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}
