package licensing

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func init() {
	conf.ContributeValidator(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		if cfg.SiteConfig().LicenseKey != "" {
			info, _, err := ParseProductLicenseKeyWithBuiltinOrGenerationKey(cfg.SiteConfig().LicenseKey)
			if err != nil {
				return conf.NewSiteProblems("should provide a valid license key")
			} else if err = info.hasUnknownPlan(); err != nil {
				return conf.NewSiteProblems(err.Error())
			}
		}
		return nil
	})
}
