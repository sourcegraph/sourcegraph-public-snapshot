package indexer

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type RankingService interface {
	RepositoryIndexer(interval time.Duration) goroutine.BackgroundRoutine
}
