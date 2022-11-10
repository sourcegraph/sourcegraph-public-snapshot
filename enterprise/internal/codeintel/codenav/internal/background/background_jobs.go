package background

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type BackgroundJob interface {
	NewRankingGraphSerializer(interval time.Duration) goroutine.BackgroundRoutine
	SetService(service CodeNavService)
}

type backgroundJob struct {
	codeNavSvc CodeNavService
	operations *operations
}

func New(
	observationContext *observation.Context,
) BackgroundJob {
	return &backgroundJob{
		operations: newOperations(observationContext),
	}
}

func (b *backgroundJob) SetService(codeNavSvc CodeNavService) {
	b.codeNavSvc = codeNavSvc
}
