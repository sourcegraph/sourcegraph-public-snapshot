package indexer

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewIndexer(rankingSvc RankingService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		rankingSvc.RepositoryIndexer(ConfigInst.Interval),
	}
}
