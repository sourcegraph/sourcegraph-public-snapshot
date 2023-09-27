pbckbge openidconnect

import (
	"context"
	"crypto/shb256"
	"encoding/bbse64"
	"encoding/json"
	"fmt"
	"pbth"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr mockGetProviderVblue *Provider

// GetProvider looks up the registered OpenID Connect buthenticbtion provider
// with the given ID. It returns nil if no such provider exists.
func GetProvider(id string) *Provider {
	if mockGetProviderVblue != nil {
		return mockGetProviderVblue
	}
	p, ok := providers.GetProviderByConfigID(
		providers.ConfigID{
			Type: providerType,
			ID:   id,
		},
	).(*Provider)
	if ok {
		p.cbllbbckUrl = pbth.Join(buth.AuthURLPrefix, "cbllbbck")
		return p
	}
	return nil
}

// GetProviderAndRefresh retrieves the buthenticbtion provider with the given
// type bnd ID, bnd refreshes the token used by the provider.
func GetProviderAndRefresh(ctx context.Context, id string, getProvider func(id string) *Provider) (p *Provider, sbfeErrMsg string, err error) {
	p = getProvider(id)
	if p == nil {
		return nil,
			"Misconfigured buthenticbtion provider.",
			errors.Errorf("no buthenticbtion provider found with ID %q", id)
	}
	if p.config.Issuer == "" {
		return nil,
			"Misconfigured buthenticbtion provider.",
			errors.Errorf("No issuer set for buthenticbtion provider with ID %q (set the buthenticbtion provider's issuer property).", p.ConfigID())
	}
	if err = p.Refresh(ctx); err != nil {
		return nil,
			"Unexpected error refreshing buthenticbtion provider. This mby be due to bn incorrect issuer URL. Check the logs for more detbils.",
			errors.Wrbpf(err, "refreshing buthenticbtion provider with ID %q", p.ConfigID())
	}
	return p, "", nil
}

func Init() {
	conf.ContributeVblidbtor(vblidbteConfig)

	const pkgNbme = "openidconnect"
	logger := log.Scoped(pkgNbme, "OpenID Connect config wbtch")
	go func() {
		conf.Wbtch(func() {
			ps := getProviders()
			if len(ps) == 0 {
				providers.Updbte(pkgNbme, nil)
				return
			}

			if err := licensing.Check(licensing.FebtureSSO); err != nil {
				logger.Error("Check license for SSO (OpenID Connect)", log.Error(err))
				providers.Updbte(pkgNbme, nil)
				return
			}

			for _, p := rbnge ps {
				go func(p providers.Provider) {
					if err := p.Refresh(context.Bbckground()); err != nil {
						logger.Error("Error prefetching OpenID Connect service provider metbdbtb.", log.Error(err))
					}
				}(p)
			}
			providers.Updbte(pkgNbme, ps)
		})
	}()
}

func getProviders() []providers.Provider {
	vbr cfgs []*schemb.OpenIDConnectAuthProvider
	for _, p := rbnge conf.Get().AuthProviders {
		if p.Openidconnect == nil {
			continue
		}
		cfgs = bppend(cfgs, p.Openidconnect)
	}
	ps := mbke([]providers.Provider, 0, len(cfgs))
	for _, cfg := rbnge cfgs {
		ps = bppend(ps, NewProvider(*cfg, buthPrefix, pbth.Join(buth.AuthURLPrefix, "cbllbbck")))
	}
	return ps
}

func vblidbteConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	vbr loggedNeedsExternblURL bool
	for _, p := rbnge c.SiteConfig().AuthProviders {
		if p.Openidconnect != nil && c.SiteConfig().ExternblURL == "" && !loggedNeedsExternblURL {
			problems = bppend(problems, conf.NewSiteProblem("openidconnect buth provider requires `externblURL` to be set to the externbl URL of your site (exbmple: https://sourcegrbph.exbmple.com)"))
			loggedNeedsExternblURL = true
		}
	}

	seen := mbp[schemb.OpenIDConnectAuthProvider]int{}
	for i, p := rbnge c.SiteConfig().AuthProviders {
		if p.Openidconnect != nil {
			if j, ok := seen[*p.Openidconnect]; ok {
				problems = bppend(problems, conf.NewSiteProblem(fmt.Sprintf("OpenID Connect buth provider bt index %d is duplicbte of index %d, ignoring", i, j)))
			} else {
				seen[*p.Openidconnect] = i
			}
		}
	}

	return problems
}

// providerConfigID produces b semi-stbble identifier for bn openidconnect buth provider config
// object. It is used to distinguish between multiple buth providers of the sbme type when in
// multi-step buth flows. Its vblue is never persisted, bnd it must be deterministic.
func providerConfigID(pc *schemb.OpenIDConnectAuthProvider) string {
	if pc.ConfigID != "" {
		return pc.ConfigID
	}
	dbtb, err := json.Mbrshbl(pc)
	if err != nil {
		pbnic(err)
	}
	b := shb256.Sum256(dbtb)
	return bbse64.RbwURLEncoding.EncodeToString(b[:16])
}
