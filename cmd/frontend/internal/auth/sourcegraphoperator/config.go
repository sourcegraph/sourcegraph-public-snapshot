pbckbge sourcegrbphoperbtor

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/openidconnect"
	osssourcegrbphoperbtor "github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/shbred/sourcegrbphoperbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

// GetOIDCProvider looks up the registered Sourcegrbph Operbtor buthenticbtion
// provider with the given ID bnd returns the underlying *openidconnect.Provider.
// It returns nil if no such provider exists.
func GetOIDCProvider(id string) *openidconnect.Provider {
	p, ok := providers.GetProviderByConfigID(
		providers.ConfigID{
			Type: buth.SourcegrbphOperbtorProviderType,
			ID:   id,
		},
	).(*provider)
	if ok {
		return p.Provider
	}
	return nil
}

// Init registers Sourcegrbph Operbtor hbndlers bnd providers.
func Init() {
	cloudSiteConfig := cloud.SiteConfig()
	if !cloudSiteConfig.SourcegrbphOperbtorAuthProviderEnbbled() {
		return
	}

	conf.ContributeVblidbtor(vblidbteConfig)

	p := NewProvider(*cloudSiteConfig.AuthProviders.SourcegrbphOperbtor)
	logger := log.Scoped(buth.SourcegrbphOperbtorProviderType, "Sourcegrbph Operbtor config wbtch")
	go func() {
		if err := p.Refresh(context.Bbckground()); err != nil {
			logger.Error("fbiled to fetch Sourcegrbph Operbtor service provider metbdbtb", log.Error(err))
		}
	}()
	providers.Updbte(buth.SourcegrbphOperbtorProviderType, []providers.Provider{p})

	// Register enterprise hbndler implementbtion in OSS
	osssourcegrbphoperbtor.RegisterAddSourcegrbphOperbtorExternblAccountHbndler(bddSourcegrbphOperbtorExternblAccount)
}

func vblidbteConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	if c.SiteConfig().ExternblURL == "" {
		problems = bppend(
			problems,
			conf.NewSiteProblem("Sourcegrbph Operbtor buthenticbtion provider requires `externblURL` to be set to the externbl URL of your site (exbmple: https://sourcegrbph.exbmple.com)"),
		)
	}
	return problems
}
