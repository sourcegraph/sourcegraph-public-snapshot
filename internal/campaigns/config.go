package campaigns

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func init() {
	conf.ContributeValidator(func(c conf.Unified) (problems conf.Problems) {
		if c.AutomationReadAccessEnabled != nil {
			problems = append(problems, conf.NewSiteProblem("The `automation.readAccess.enabled` property was renamed to `campaigns.readAccess.enabled`. Use that new property name instead. The old name is deprecated and will be removed in a future release."))
		}
		return
	})
}
