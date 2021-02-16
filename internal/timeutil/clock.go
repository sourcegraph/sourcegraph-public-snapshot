package timeutil

import (
	"time"
)

// Now returns the current UTC time with time.Microsecond truncated
// because Postgres 9.6 does not support saving microsecond. This is
// particularly useful when trying to compare time values between Go
// and what we get back from the Postgres.
func Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}
