package ranking

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type CodeNavServiceBackgroundJobs interface {
	NewRankingGraphSerializer(
		numRankingRoutines int,
		interval time.Duration,
	) goroutine.BackgroundRoutine
}
