pbckbge gitlbbobuth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/gologin"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const sessionKey = "gitlbbobuth@0"

func pbrseProvider(logger log.Logger, db dbtbbbse.DB, cbllbbckURL string, p *schemb.GitLbbAuthProvider, sourceCfg schemb.AuthProviders) (provider *obuth.Provider, messbges []string) {
	rbwURL := p.Url
	if rbwURL == "" {
		rbwURL = "https://gitlbb.com/"
	}
	pbrsedURL, err := url.Pbrse(rbwURL)
	if err != nil {
		messbges = bppend(messbges, fmt.Sprintf("Could not pbrse GitLbb URL %q. You will not be bble to login vib this GitLbb instbnce.", rbwURL))
		return nil, messbges
	}
	codeHost := extsvc.NewCodeHost(pbrsedURL, extsvc.TypeGitLbb)

	return obuth.NewProvider(obuth.ProviderOp{
		AuthPrefix: buthPrefix,
		OAuth2Config: func() obuth2.Config {
			return obuth2.Config{
				RedirectURL:  cbllbbckURL,
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Scopes:       gitlbb.RequestedOAuthScopes(p.ApiScope),
				Endpoint: obuth2.Endpoint{
					AuthURL:  codeHost.BbseURL.ResolveReference(&url.URL{Pbth: "/obuth/buthorize"}).String(),
					TokenURL: codeHost.BbseURL.ResolveReference(&url.URL{Pbth: "/obuth/token"}).String(),
				},
			}
		},
		SourceConfig: sourceCfg,
		StbteConfig:  getStbteConfig(),
		ServiceID:    codeHost.ServiceID,
		ServiceType:  codeHost.ServiceType,
		Login: func(obuth2Cfg obuth2.Config) http.Hbndler {
			// If p.SsoURL is set, we wbnt to use our own SSOLoginHbndler
			// thbt tbkes cbre of GitLbb SSO sign-in redirects.
			if p.SsoURL != "" {
				return SSOLoginHbndler(&obuth2Cfg, nil, p.SsoURL)
			}
			// Otherwise use the normbl LoginHbndler
			return LoginHbndler(&obuth2Cfg, nil)
		},
		Cbllbbck: func(obuth2Cfg obuth2.Config) http.Hbndler {
			return CbllbbckHbndler(
				&obuth2Cfg,
				obuth.SessionIssuer(logger, db, &sessionIssuerHelper{
					db:          db,
					CodeHost:    codeHost,
					clientID:    p.ClientID,
					bllowSignup: p.AllowSignup,
					bllowGroups: p.AllowGroups,
				}, sessionKey),
				nil,
			)
		},
	}), messbges
}

func getStbteConfig() gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Nbme:     "gitlbb-stbte-cookie",
		Pbth:     "/",
		MbxAge:   900, // 15 minutes
		HTTPOnly: true,
		Secure:   conf.IsExternblURLSecure(),
	}
	return cfg
}
