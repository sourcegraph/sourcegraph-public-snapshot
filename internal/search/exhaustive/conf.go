package exhaustive

import (
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func IsEnabled(cfg conftypes.SiteConfigQuerier) bool {
	experimentalFeatures := cfg.SiteConfig().ExperimentalFeatures
	if experimentalFeatures != nil && experimentalFeatures.SearchJobs != nil {
		return *experimentalFeatures.SearchJobs
	}
	return true
}
