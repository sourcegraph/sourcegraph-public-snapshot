package ranking

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewGraphSerializers(backgroundJobs CodeNavServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewRankingGraphSerializer(ConfigInst.NumRankingRoutines, ConfigInst.RankingInterval),
	}
}
