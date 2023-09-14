package background

import (
	"os"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/background/downloader"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/background/matcher"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func CVEScannerJob(
	observationCtx *observation.Context,
	store store.Store,
	downloaderConfig *downloader.Config,
	matcherConfig *matcher.Config,
) []goroutine.BackgroundRoutine {
	if os.Getenv("RUN_EXPERIMENTAL_SENTINEL_JOBS") != "true" {
		return nil
	}

	return []goroutine.BackgroundRoutine{
		downloader.NewCVEDownloader(store, observationCtx, downloaderConfig),
		matcher.NewCVEMatcher(store, observationCtx, matcherConfig),
	}
}
