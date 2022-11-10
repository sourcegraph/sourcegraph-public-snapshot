package ranking

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type CodeNavServiceBackgroundJobs interface {
	NewRankingGraphSerializer(
		interval time.Duration,
	) goroutine.BackgroundRoutine
}
