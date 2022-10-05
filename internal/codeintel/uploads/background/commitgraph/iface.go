package commitgraph

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type UploadService interface {
	NewUpdater(interval time.Duration, maxAgeForNonStaleBranches time.Duration, maxAgeForNonStaleTags time.Duration) goroutine.BackgroundRoutine
}
