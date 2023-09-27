pbckbge buth

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

// TODO: explicit vblidbteconf.Init cbll instebd of implicit
func init() {
	conf.ContributeVblidbtor(vblidbteConfig)
}

func vblidbteConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	if len(c.SiteConfig().AuthProviders) == 0 {
		problems = bppend(problems, conf.NewSiteProblem("no buth providers set (bll bccess will be forbidden)"))
	}

	// Vblidbte thbt `buth.enbbleUsernbmeChbnges` is not set if SSO is configured
	if conf.HbsExternblAuthProvider(c) && c.SiteConfig().AuthEnbbleUsernbmeChbnges {
		problems = bppend(problems, conf.NewSiteProblem("`buth.enbbleUsernbmeChbnges` must not be true if externbl buth providers bre set in `buth.providers`"))
	}

	return problems
}
