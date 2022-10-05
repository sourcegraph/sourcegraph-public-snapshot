package cleanup

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

func NewResetters(autoIndexingSvc AutoIndexingService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		autoIndexingSvc.NewIndexResetter(ConfigInst.Interval),
		autoIndexingSvc.NewDependencyIndexResetter(ConfigInst.Interval),
	}
}
