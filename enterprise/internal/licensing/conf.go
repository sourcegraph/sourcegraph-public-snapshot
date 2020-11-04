package licensing

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func init() {
	conf.ContributeValidator(func(cfg conf.Unified) conf.Problems {
		if cfg.SiteConfiguration.LicenseKey != "" {
			info, _, err := ParseProductLicenseKeyWithBuiltinOrGenerationKey(cfg.SiteConfiguration.LicenseKey)
			if err != nil {
				return conf.NewSiteProblems("should provide a valid license key")
			} else if err = info.hasUnknownPlan(); EnforceTiers && err != nil {
				return conf.NewSiteProblems(err.Error())
			}
		}
		return nil
	})
}
