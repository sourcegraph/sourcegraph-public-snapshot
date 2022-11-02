package loader

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewPageRankLoader(rankingSvc RankingService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		rankingSvc.RankLoader(ConfigInst.LoadInterval),
		rankingSvc.RankMerger(ConfigInst.MergeInterval),
	}
}
