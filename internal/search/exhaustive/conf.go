package exhaustive

import (
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
)

func IsEnabled(cfg conftypes.SiteConfigQuerier) bool {
	if dotcom.SourcegraphDotComMode() {
		return false
	}

	// TODO(stefan): Remove this once Search Jobs is no longer experimental.
	experimentalFeatures := cfg.SiteConfig().ExperimentalFeatures
	if experimentalFeatures != nil && experimentalFeatures.SearchJobs != nil {
		return *experimentalFeatures.SearchJobs
	}

	if v, _ := strconv.ParseBool(os.Getenv("DISABLE_SEARCH_JOBS")); v {
		return false
	}

	return true
}
