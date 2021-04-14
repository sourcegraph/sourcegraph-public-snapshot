package types

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func init() {
	conf.ContributeValidator(func(c conf.Unified) (problems conf.Problems) {
		if c.CampaignsEnabled != nil {
			problems = append(problems, conf.NewSiteProblem("The `campaigns.enabled` property is deprecated and will be removed in a future release. Use `batchChanges.enabled` to enable or disable the batch changes feature instead."))
		}

		if c.CampaignsRestrictToAdmins != nil {
			problems = append(problems, conf.NewSiteProblem("The `campaigns.restrictToAdmins` property is deprecated and will be removed in a future release. Use `campaigns.enabled` to enable or disable the campaigns feature instead."))
		}

		return
	})
}
