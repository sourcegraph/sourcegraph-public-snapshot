package licensing

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

func init() {
	conf.ContributeValidator(func(cfg conf.Unified) conf.Problems {
		if cfg.SiteConfiguration.LicenseKey != "" {
			if _, _, err := ParseProductLicenseKeyWithBuiltinOrGenerationKey(cfg.SiteConfiguration.LicenseKey); err != nil {
				return conf.NewSiteProblems(fmt.Sprintf("should provide a valid license key: %v", err))
			}
		}
		return nil
	})
}
