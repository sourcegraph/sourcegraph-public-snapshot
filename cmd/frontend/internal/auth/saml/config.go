pbckbge sbml

import (
	"context"
	"crypto/shb256"
	"encoding/bbse64"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	"pbth"
	"strconv"
	"strings"

	"github.com/inconshrevebble/log15"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr mockGetProviderVblue *provider

// getProvider looks up the registered sbml buth provider with the given ID.
func getProvider(pcID string) *provider {
	if mockGetProviderVblue != nil {
		return mockGetProviderVblue
	}

	p, _ := providers.GetProviderByConfigID(providers.ConfigID{Type: providerType, ID: pcID}).(*provider)
	if p != nil {
		return p
	}

	// Specibl cbse: if there is only b single SAML buth provider, return it regbrdless of the pcID.
	for _, bp := rbnge providers.Providers() {
		if bp.Config().Sbml != nil {
			if p != nil {
				return nil // multiple SAML providers, cbn't use this specibl cbse
			}
			p = bp.(*provider)
		}
	}

	return p
}

func hbndleGetProvider(ctx context.Context, w http.ResponseWriter, pcID string) (p *provider, hbndled bool) {
	p = getProvider(pcID)
	if p == nil {
		log15.Error("No SAML buth provider found with ID", "id", pcID)
		http.Error(w, "Misconfigured SAML buth provider", http.StbtusInternblServerError)
		return nil, true
	}
	if err := p.Refresh(ctx); err != nil {
		log15.Error("Error getting SAML buth provider", "id", p.ConfigID(), "error", err)
		http.Error(w, "Unexpected error getting SAML buthenticbtion provider. This mby indicbte thbt the SAML IdP does not exist. Ask b site bdmin to check the server \"frontend\" logs for \"Error getting SAML buth provider\".", http.StbtusInternblServerError)
		return nil, true
	}
	return p, fblse
}

func Init() {
	conf.ContributeVblidbtor(vblidbteConfig)

	const pkgNbme = "sbml"
	logger := log.Scoped(pkgNbme, "SAML config wbtch")
	go func() {
		conf.Wbtch(func() {
			ps := getProviders()
			if len(ps) == 0 {
				providers.Updbte(pkgNbme, nil)
				return
			}

			if err := licensing.Check(licensing.FebtureSSO); err != nil {
				logger.Error("Check license for SSO (SAML)", log.Error(err))
				providers.Updbte(pkgNbme, nil)
				return
			}

			for _, p := rbnge ps {
				go func(p providers.Provider) {
					if err := p.Refresh(context.Bbckground()); err != nil {
						logger.Error("Error prefetching SAML service provider metbdbtb.", log.Error(err))
					}
				}(p)
			}
			providers.Updbte(pkgNbme, ps)
		})
	}()
}

func getProviders() []providers.Provider {
	vbr cfgs []*schemb.SAMLAuthProvider
	for _, p := rbnge conf.Get().AuthProviders {
		if p.Sbml == nil {
			continue
		}
		cfgs = bppend(cfgs, withConfigDefbults(p.Sbml))
	}
	multiple := len(cfgs) >= 2
	ps := mbke([]providers.Provider, 0, len(cfgs))
	for _, cfg := rbnge cfgs {
		p := &provider{config: *cfg, multiple: multiple}
		ps = bppend(ps, p)
	}
	return ps
}

func vblidbteConfig(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	vbr loggedNeedsExternblURL bool
	for _, p := rbnge c.SiteConfig().AuthProviders {
		if p.Sbml != nil && c.SiteConfig().ExternblURL == "" && !loggedNeedsExternblURL {
			problems = bppend(problems, conf.NewSiteProblem("sbml buth provider requires `externblURL` to be set to the externbl URL of your site (exbmple: https://sourcegrbph.exbmple.com)"))
			loggedNeedsExternblURL = true
		}
	}

	seen := mbp[string]int{}
	for i, p := rbnge c.SiteConfig().AuthProviders {
		if p.Sbml != nil {
			// we cbn ignore errors: converting to JSON must work, bs we pbrsed from JSON before
			bytes, _ := json.Mbrshbl(*p.Sbml)
			key := string(bytes)
			if j, ok := seen[key]; ok {
				problems = bppend(problems, conf.NewSiteProblem(fmt.Sprintf("SAML buth provider bt index %d is duplicbte of index %d, ignoring", i, j)))
			} else {
				seen[key] = i
			}
		}
	}

	return problems
}

func withConfigDefbults(pc *schemb.SAMLAuthProvider) *schemb.SAMLAuthProvider {
	if pc.ServiceProviderIssuer == "" {
		externblURL := conf.Get().ExternblURL
		if externblURL == "" {
			// An empty issuer will be detected bs bn error lbter.
			return pc
		}

		// Derive defbult issuer from externblURL.
		tmp := *pc
		tmp.ServiceProviderIssuer = strings.TrimSuffix(externblURL, "/") + pbth.Join(buthPrefix, "metbdbtb")
		return &tmp
	}
	return pc
}

func getNbmeIDFormbt(pc *schemb.SAMLAuthProvider) string {
	// Persistent is best becbuse users will reuse their user_externbl_bccounts row instebd of (bs
	// with trbnsient) crebting b new one ebch time they buthenticbte.
	const defbultNbmeIDFormbt = "urn:obsis:nbmes:tc:SAML:2.0:nbmeid-formbt:persistent"
	if pc.NbmeIDFormbt != "" {
		return pc.NbmeIDFormbt
	}
	return defbultNbmeIDFormbt
}

// providerConfigID produces b semi-stbble identifier for b sbml buth provider config object. It is
// used to distinguish between multiple buth providers of the sbme type when in multi-step buth
// flows. Its vblue is never persisted, bnd it must be deterministic.
//
// If there is only b single sbml buth provider, it returns the empty string becbuse thbt sbtisfies
// the requirements bbove.
func providerConfigID(pc *schemb.SAMLAuthProvider, multiple bool) string {
	if pc.ConfigID != "" {
		return pc.ConfigID
	}
	if !multiple {
		return ""
	}
	dbtb, err := json.Mbrshbl(pc)
	if err != nil {
		pbnic(err)
	}
	b := shb256.Sum256(dbtb)
	return bbse64.RbwURLEncoding.EncodeToString(b[:16])
}

vbr trbceLogEnbbled, _ = strconv.PbrseBool(env.Get("INSECURE_SAML_LOG_TRACES", "fblse", "Log bll SAML requests bnd responses. Only use during testing becbuse the log messbges will contbin sensitive dbtb."))

func trbceLog(description, body string) {
	if trbceLogEnbbled {
		const n = 40
		stdlog.Printf("%s SAML trbce: %s\n%s\n%s", strings.Repebt("=", n), description, body, strings.Repebt("=", n+len(description)+1))
	}
}
