package check

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Any(checks ...CheckFunc) CheckFunc {
	return func(ctx context.Context) (err error) {
		for _, chk := range checks {
			err = chk(ctx)
			if err == nil {
				return nil
			}
		}
		return err
	}
}

func Combine(checks ...CheckFunc) CheckFunc {
	return func(ctx context.Context) (err error) {
		for _, chk := range checks {
			err = chk(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func Retry(check CheckFunc, retries int, sleep time.Duration) CheckFunc {
	return func(ctx context.Context) (err error) {
		for i := 0; i < retries; i++ {
			err = check(ctx)
			if err == nil {
				return nil
			}
			time.Sleep(sleep)
		}
		return err
	}
}

func WrapErrMessage(check CheckFunc, message string) CheckFunc {
	return func(ctx context.Context) error {
		err := check(ctx)
		if err != nil {
			return errors.Wrap(err, message)
		}
		return nil
	}
}
