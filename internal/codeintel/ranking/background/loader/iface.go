package loader

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type RankingService interface {
	RankLoader(interval time.Duration) goroutine.BackgroundRoutine
	RankMerger(interval time.Duration) goroutine.BackgroundRoutine
}
