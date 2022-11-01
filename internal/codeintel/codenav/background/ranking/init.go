package ranking

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewGraphSerializers(backgroundJobs CodeNavServiceBackgroundJobs) (routines []goroutine.BackgroundRoutine) {
	for i := 0; i < ConfigInst.NumRankingRoutines; i++ {
		routines = append(routines, backgroundJobs.NewRankingGraphSerializer(ConfigInst.RankingInterval))
	}

	return routines
}
