package batches

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/scheduler/window"
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func init() {
	conf.ContributeValidator(func(c conf.Unified) (problems conf.Problems) {
		if err := window.ValidateConfiguration(c.BatchChangesRolloutWindows); err != nil {
			problems = append(problems, conf.NewSiteProblem(err.Error()))
		}

		return
	})
}
