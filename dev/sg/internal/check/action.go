package check

import (
	"context"
)

// CheckAction adapts simple CheckFuncs into the more complex ActionFunc interface.
func CheckAction[Args any](fn CheckFunc) ActionFunc[Args] {
	return func(ctx context.Context, cio IO, args Args) error {
		return fn(ctx)
	}
}
