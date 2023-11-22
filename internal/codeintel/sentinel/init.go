package sentinel

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/background/downloader"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/background/matcher"
	sentinelstore "github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewService(
	observationCtx *observation.Context,
	db database.DB,
) *Service {
	return newService(
		scopedContext("service", observationCtx),
		sentinelstore.New(scopedContext("store", observationCtx), db),
	)
}

var (
	DownloaderConfigInst = &downloader.Config{}
	MatcherConfigInst    = &matcher.Config{}
)

func CVEScannerJob(observationCtx *observation.Context, service *Service) []goroutine.BackgroundRoutine {
	return background.CVEScannerJob(
		scopedContext("cvescanner", observationCtx),
		service.store,
		DownloaderConfigInst,
		MatcherConfigInst,
	)
}

func scopedContext(component string, parent *observation.Context) *observation.Context {
	return observation.ScopedContext("codeintel", "sentinel", component, parent)
}
