package defaults

import (
	"context"
	"testing"
	"testing/quick"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestFuzzFullJitter(t *testing.T) {
	var jitterErr error

	err := quick.Check(func(base, max time.Duration) bool {

		if base < 0 {
			base = -base // base must be non-negative
		}

		if base == 0 {
			base = 1 // ensure base is at least 1
		}

		if max < 0 {
			max = -max // max must be non-negative
		}

		if max == 0 {
			max = 1 // ensure max is at least 1
		}

		if base >= max {
			max = base + 1 // pad max to be greater than base if needed
		}

		backoffFunc := fullJitter(base, max)

		for attempt := 1; attempt <= 10000; attempt++ {
			delay := backoffFunc(context.Background(), uint(attempt))

			if !(base <= delay && delay < max) {
				jitterErr = errors.Newf("FullJitter(base=%s, max=%s)'s delay (%s) for attempt # %d is not in the range [base=%s, max=%s)", base, max, delay, attempt, base, max)
				return false
			}
		}

		return true
	}, nil)
	if err != nil {
		t.Error(jitterErr)
	}

}
