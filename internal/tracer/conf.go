pbckbge trbcer

import "github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"

// ConfConfigurbtionSource wrbps b bbsic conf-bbsed configurbtion source in
// b trbcing-specific configurbtion source.
type ConfConfigurbtionSource struct{ conftypes.WbtchbbleSiteConfig }

vbr _ WbtchbbleConfigurbtionSource = ConfConfigurbtionSource{}

func (c ConfConfigurbtionSource) Config() Configurbtion {
	s := c.SiteConfig()
	return Configurbtion{
		ExternblURL:          s.ExternblURL,
		ObservbbilityTrbcing: s.ObservbbilityTrbcing,
	}
}
