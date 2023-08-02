package confvalidators

import (
	"github.com/sourcegraph/sourcegraph/internal/batches/types/scheduler/window"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func validateBatchChangeRolloutWindows(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	if _, err := window.NewConfiguration(c.SiteConfig().BatchChangesRolloutWindows); err != nil {
		problems = append(problems, conf.NewSiteProblem(err.Error()))
	}

	return problems
}
