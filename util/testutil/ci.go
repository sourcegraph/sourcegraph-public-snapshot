package testutil

import (
	"os"
	"time"
)

// Allow more time in CI because CI machines are typically
// slower and false negatives are more annoying.
var ciFactor = func() time.Duration {
	if os.Getenv("CI") == "" {
		return 1
	}
	return 3
}()
