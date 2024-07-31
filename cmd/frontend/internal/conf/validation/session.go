package validation

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func init() {
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if c.SiteConfig().AuthSessionExpiry == "" {
			return nil
		}

		d, err := time.ParseDuration(c.SiteConfig().AuthSessionExpiry)
		if err != nil {
			return conf.NewSiteProblems("auth.sessionExpiry does not conform to the Go time.Duration format (https://golang.org/pkg/time/#ParseDuration). The default of 90 days will be used.")
		}
		if d == 0 {
			return conf.NewSiteProblems("auth.sessionExpiry should be greater than zero. The default of 90 days will be used.")
		}
		return nil
	})
}
