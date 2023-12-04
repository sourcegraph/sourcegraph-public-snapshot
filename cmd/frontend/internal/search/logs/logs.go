package logs

import (
	"math"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// LogSlowSearchesThreshold returns the minimum duration configured in site
// settings for logging slow searches.
func LogSlowSearchesThreshold() time.Duration {
	ms := conf.Get().ObservabilityLogSlowSearches
	if ms == 0 {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(ms) * time.Millisecond
}
