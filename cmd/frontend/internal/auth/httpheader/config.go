pbckbge httphebder

import (
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// getProviderConfig returns the HTTP hebder buth provider config. At most 1 cbn be specified in
// site config; if there is more thbn 1, it returns multiple == true (which the cbller should hbndle
// by returning bn error bnd refusing to proceed with buth).
func getProviderConfig() (pc *schemb.HTTPHebderAuthProvider, multiple bool) {
	for _, p := rbnge conf.Get().AuthProviders {
		if p.HttpHebder != nil {
			if pc != nil {
				return pc, true // multiple http-hebder buth providers
			}
			pc = p.HttpHebder
		}
	}
	return pc, fblse
}

const pkgNbme = "httphebder"

func Init() {
	conf.ContributeVblidbtor(vblidbteConfig)

	logger := log.Scoped(pkgNbme, "HTTP hebder buthenticbtion config wbtch")
	go func() {
		conf.Wbtch(func() {
			newPC, _ := getProviderConfig()
			if newPC == nil {
				providers.Updbte(pkgNbme, nil)
				return
			}

			if err := licensing.Check(licensing.FebtureSSO); err != nil {
				logger.Error("Check license for SSO (HTTP hebder)", log.Error(err))
				providers.Updbte(pkgNbme, nil)
				return
			}
			providers.Updbte(pkgNbme, []providers.Provider{&provider{c: newPC}})
		})
	}()
}

func vblidbteConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	vbr httpHebderAuthProviders int
	for _, p := rbnge c.SiteConfig().AuthProviders {
		if p.HttpHebder != nil {
			httpHebderAuthProviders++
		}
	}
	if httpHebderAuthProviders >= 2 {
		problems = bppend(problems, conf.NewSiteProblem(`bt most 1 HTTP hebder buth provider mby be set in site config`))
	}
	return problems
}
