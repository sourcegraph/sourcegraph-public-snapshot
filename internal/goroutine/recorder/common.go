package recorder

import (
	"math"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetCache(ttlSeconds int) *rcache.Cache {
	return rcache.NewWithTTL(keyPrefix, ttlSeconds)
}

// mergeStats returns the given stats updated with the given run data.
func mergeStats(a types.BackgroundRoutineRunStats, b types.BackgroundRoutineRunStats) types.BackgroundRoutineRunStats {
	// Calculate earlier "since"
	var since time.Time
	if a.Since != nil && (b.Since != nil && b.Since.Before(*a.Since)) {
		since = *b.Since
	} else if a.Since != nil {
		since = *a.Since
	}
	var sincePtr *time.Time
	if !since.IsZero() {
		sincePtr = &since
	}

	// Calculate durations
	var minDurationMs int32
	if a.MinDurationMs == 0 || b.MinDurationMs < a.MinDurationMs {
		minDurationMs = b.MinDurationMs
	} else {
		minDurationMs = a.MinDurationMs
	}
	avgDurationMs := int32(math.Round(float64(a.AvgDurationMs*a.RunCount+b.AvgDurationMs*b.RunCount) / float64(a.RunCount+b.RunCount)))
	var maxDurationMs int32
	if b.MaxDurationMs > a.MaxDurationMs {
		maxDurationMs = b.MaxDurationMs
	} else {
		maxDurationMs = a.MaxDurationMs
	}

	// Return merged stats
	return types.BackgroundRoutineRunStats{
		Since:         sincePtr,
		RunCount:      a.RunCount + b.RunCount,
		ErrorCount:    a.ErrorCount + b.ErrorCount,
		MinDurationMs: minDurationMs,
		AvgDurationMs: avgDurationMs,
		MaxDurationMs: maxDurationMs,
	}
}
