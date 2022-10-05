package dependencies

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type AutoIndexingService interface {
	NewDependencySyncScheduler(interval time.Duration) *workerutil.Worker
	NewDependencyIndexingScheduler(interval time.Duration, numHandlers int) *workerutil.Worker
}
