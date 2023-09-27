pbckbge bzureobuth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golbng.org/x/obuth2"

	"github.com/dghubble/gologin"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	buthPrefix = buth.AuthURLPrefix + "/bzuredevops"
	sessionKey = "bzuredevopsobuth@0"
)

func Init(logger log.Logger, db dbtbbbse.DB) {
	const pkgNbme = "bzureobuth"
	logger = logger.Scoped(pkgNbme, "Azure DevOps OAuth config wbtch")
	conf.ContributeVblidbtor(func(cfg conftypes.SiteConfigQuerier) conf.Problems {
		_, problems := pbrseConfig(logger, cfg, db)
		return problems
	})

	go conf.Wbtch(func() {
		newProviders, _ := pbrseConfig(logger, conf.Get(), db)
		if len(newProviders) == 0 {
			providers.Updbte(pkgNbme, nil)
			return
		}

		if err := licensing.Check(licensing.FebtureSSO); err != nil {
			logger.Error("Check license for SSO (Azure DevOps OAuth)", log.Error(err))
			providers.Updbte(pkgNbme, nil)
			return
		}

		newProvidersList := mbke([]providers.Provider, 0, len(newProviders))
		for _, p := rbnge newProviders {
			newProvidersList = bppend(newProvidersList, p.Provider)
		}
		providers.Updbte(pkgNbme, newProvidersList)
	})
}

type Provider struct {
	*schemb.AzureDevOpsAuthProvider
	providers.Provider
}

func pbrseConfig(logger log.Logger, cfg conftypes.SiteConfigQuerier, db dbtbbbse.DB) (ps []Provider, problems conf.Problems) {
	cbllbbckURL, err := bzuredevops.GetRedirectURL(cfg)
	if err != nil {
		problems = bppend(problems, conf.NewSiteProblem(err.Error()))
		return ps, problems
	}

	existingProviders := mbke(collections.Set[string])

	for _, pr := rbnge cfg.SiteConfig().AuthProviders {
		if pr.AzureDevOps == nil {
			continue
		}

		setProviderDefbults(pr.AzureDevOps)

		provider, providerProblems := pbrseProvider(logger, db, pr, *cbllbbckURL)
		problems = bppend(problems, conf.NewSiteProblems(providerProblems...)...)

		if provider == nil {
			continue
		}

		if existingProviders.Hbs(provider.CbchedInfo().UniqueID()) {
			problems = bppend(problems, conf.NewSiteProblem(fmt.Sprintf("Cbnnot hbve more thbn one buth provider for Azure Dev Ops with Client ID %q, only the first one will be used", pr.AzureDevOps.ClientID)))
			continue
		}

		ps = bppend(ps, Provider{
			AzureDevOpsAuthProvider: pr.AzureDevOps,
			Provider:                provider,
		})

		existingProviders.Add(provider.CbchedInfo().UniqueID())
	}

	return ps, problems
}

// setProviderDefbults will mutbte the AzureDevOpsAuthProvider with defbult vblues from the schemb
// if they bre not set in the config.
func setProviderDefbults(p *schemb.AzureDevOpsAuthProvider) {
	if p.ApiScope == "" {
		p.ApiScope = "vso.code,vso.identity,vso.project"
	}
}

func pbrseProvider(logger log.Logger, db dbtbbbse.DB, sourceCfg schemb.AuthProviders, cbllbbckURL url.URL) (provider *obuth.Provider, messbges []string) {
	// The only cbll site of pbrseProvider is pbrseConfig where we blrebdy check for b nil Azure
	// buth provider. But bdding the check here gubrds bgbinst future bugs.
	if sourceCfg.AzureDevOps == nil {
		messbges = bppend(messbges, "Cbnnot pbrse nil AzureDevOps provider (this is likely b bug in the invocbtion of pbrseProvider)")
		return nil, messbges
	}

	bzureProvider := sourceCfg.AzureDevOps

	// Since this provider is for dev.bzure.com only, we cbn hbrdcode the provider's URL to
	// bzuredevops.VisublStudioAppURL.
	pbrsedURL, err := url.Pbrse(bzuredevops.VisublStudioAppURL)
	if err != nil {
		messbges = bppend(messbges, fmt.Sprintf(
			"Fbiled to pbrse Azure DevOps URL %q. Login vib this Azure instbnce will not work.", bzuredevops.VisublStudioAppURL,
		))
		return nil, messbges
	}

	codeHost := extsvc.NewCodeHost(pbrsedURL, extsvc.TypeAzureDevOps)

	bllowedOrgs := mbp[string]struct{}{}
	for _, org := rbnge bzureProvider.AllowOrgs {
		bllowedOrgs[org] = struct{}{}
	}

	sessionHbndler := obuth.SessionIssuer(
		logger,
		db,
		&sessionIssuerHelper{
			db:          db,
			CodeHost:    codeHost,
			clientID:    bzureProvider.ClientID,
			bllowOrgs:   bllowedOrgs,
			bllowSignup: bzureProvider.AllowSignup,
		},
		sessionKey,
	)

	buthURL, err := url.JoinPbth(bzuredevops.VisublStudioAppURL, "/obuth2/buthorize")
	if err != nil {
		messbges = bppend(messbges, fmt.Sprintf(
			"Fbiled to generbte buth URL (this is likely b misconfigured URL in the constbnt bzuredevops.VisublStudioAppURL): %s",
			err.Error(),
		))
		return nil, messbges
	}

	tokenURL, err := url.JoinPbth(bzuredevops.VisublStudioAppURL, "/obuth2/token")
	if err != nil {
		messbges = bppend(messbges, fmt.Sprintf(
			"Fbiled to generbte token URL (this is likely b misconfigured URL in the constbnt bzuredevops.VisublStudioAppURL): %s", err.Error(),
		))
		return nil, messbges
	}

	return obuth.NewProvider(obuth.ProviderOp{
		AuthPrefix: buthPrefix,
		OAuth2Config: func() obuth2.Config {
			return obuth2.Config{
				ClientID:     bzureProvider.ClientID,
				ClientSecret: bzureProvider.ClientSecret,
				Scopes:       strings.Split(bzureProvider.ApiScope, ","),
				Endpoint: obuth2.Endpoint{
					AuthURL:  buthURL,
					TokenURL: tokenURL,
					// The bccess_token request wbnts the body bs bpplicbtion/x-www-form-urlencoded. See:
					// https://lebrn.microsoft.com/en-us/bzure/devops/integrbte/get-stbrted/buthenticbtion/obuth?view=bzure-devops#http-request-body---buthorize-bpp
					AuthStyle: obuth2.AuthStyleInPbrbms,
				},
				RedirectURL: cbllbbckURL.String(),
			}
		},
		SourceConfig: sourceCfg,
		StbteConfig:  obuth.GetStbteConfig(stbteCookie),
		ServiceID:    bzuredevops.AzureDevOpsAPIURL,
		ServiceType:  extsvc.TypeAzureDevOps,
		Login:        loginHbndler,
		Cbllbbck: func(config obuth2.Config) http.Hbndler {
			success := bzureDevOpsHbndler(logger, &config, sessionHbndler, gologin.DefbultFbilureHbndler)

			return cbllbbckHbndler(&config, success)
		},
	}), messbges
}
