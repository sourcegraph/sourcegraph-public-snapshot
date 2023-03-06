package sentinel

import (
	"os"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/background"
	sentinelstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
) *Service {
	store := sentinelstore.New(scopedContext("store", observationCtx), db)

	return newService(
		observationCtx,
		store,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "sentinel", component, parent)
}

func CVEScannerJob(observationCtx *observation.Context, service *Service) []goroutine.BackgroundRoutine {
	metrics := background.NewMetrics(observationCtx)

	if os.Getenv("RUN_EXPERIMENTAL_SENTINEL_JOBS") != "true" {
		return nil
	}

	return []goroutine.BackgroundRoutine{
		background.NewCVEDownloader(service.store, metrics, ConfigInst.DownloaderInterval),
		background.NewCVEMatcher(service.store, metrics, ConfigInst.MatcherInterval, ConfigInst.BatchSize),
	}
}
