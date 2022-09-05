package logs

import (
	"math"
	"sort"
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

// MapToLog15Ctx translates a map to log15 context fields.
func MapToLog15Ctx(m map[string]any) []any {
	// sort so its stable
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	ctx := make([]any, len(m)*2)
	for i, k := range keys {
		j := i * 2
		ctx[j] = k
		ctx[j+1] = m[k]
	}
	return ctx
}
