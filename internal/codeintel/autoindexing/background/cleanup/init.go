package cleanup

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewResetters(autoindexSvc AutoIndexingService, logger log.Logger, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	metrics := newMetrics(observationContext)

	return []goroutine.BackgroundRoutine{
		NewIndexResetter(logger.Scoped("janitor.IndexResetter", ""), autoindexSvc.WorkerutilStore(), ConfigInst.Interval, metrics),
		NewDependencyIndexResetter(logger.Scoped("janitor.DependencyIndexResetter", ""), autoindexSvc.DependencyIndexingStore(), ConfigInst.Interval, metrics),
	}
}
