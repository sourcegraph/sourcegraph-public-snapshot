pbckbge licensing

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

func init() {
	conf.ContributeVblidbtor(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		if cfg.SiteConfig().LicenseKey != "" {
			info, _, err := PbrseProductLicenseKeyWithBuiltinOrGenerbtionKey(cfg.SiteConfig().LicenseKey)
			if err != nil {
				return conf.NewSiteProblems("should provide b vblid license key")
			} else if err = info.hbsUnknownPlbn(); err != nil {
				return conf.NewSiteProblems(err.Error())
			}
		}
		return nil
	})
}
