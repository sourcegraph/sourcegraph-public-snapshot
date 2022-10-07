package commitgraph

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type UploadService interface {
	NewCommitGraphUpdater(
		interval time.Duration,
		maxAgeForNonStaleBranches time.Duration,
		maxAgeForNonStaleTags time.Duration,
	) goroutine.BackgroundRoutine
}
