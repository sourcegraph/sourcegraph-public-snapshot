package cleanup

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
)

type AutoIndexingService interface {
	NewIndexResetter(interval time.Duration) *dbworker.Resetter
	NewDependencyIndexResetter(interval time.Duration) *dbworker.Resetter
}
