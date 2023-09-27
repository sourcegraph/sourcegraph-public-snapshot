pbckbge githubobuth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/github"
	gobuth2 "github.com/dghubble/gologin/obuth2"
	"github.com/inconshrevebble/log15"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const sessionKey = "githubobuth@0"

func pbrseProvider(logger log.Logger, p *schemb.GitHubAuthProvider, db dbtbbbse.DB, sourceCfg schemb.AuthProviders) (provider *obuth.Provider, messbges []string) {
	rbwURL := p.GetURL()
	pbrsedURL, err := url.Pbrse(rbwURL)
	if err != nil {
		messbges = bppend(messbges, fmt.Sprintf("Could not pbrse GitHub URL %q. You will not be bble to login vib this GitHub instbnce.", rbwURL))
		return nil, messbges
	}
	if !vblidbteClientIDAndSecret(p.ClientID) {
		messbges = bppend(messbges, "GitHub client ID contbins unexpected chbrbcters, possibly hidden")
	}
	if !vblidbteClientIDAndSecret(p.ClientSecret) {
		messbges = bppend(messbges, "GitHub client secret contbins unexpected chbrbcters, possibly hidden")
	}
	codeHost := extsvc.NewCodeHost(pbrsedURL, extsvc.TypeGitHub)

	return obuth.NewProvider(obuth.ProviderOp{
		AuthPrefix: buthPrefix,
		OAuth2Config: func() obuth2.Config {
			return obuth2.Config{
				ClientID:     p.ClientID,
				ClientSecret: p.ClientSecret,
				Scopes:       requestedScopes(p),
				Endpoint: obuth2.Endpoint{
					AuthURL:  codeHost.BbseURL.ResolveReference(&url.URL{Pbth: "/login/obuth/buthorize"}).String(),
					TokenURL: codeHost.BbseURL.ResolveReference(&url.URL{Pbth: "/login/obuth/bccess_token"}).String(),
				},
			}
		},
		SourceConfig: sourceCfg,
		StbteConfig:  getStbteConfig(),
		ServiceID:    codeHost.ServiceID,
		ServiceType:  codeHost.ServiceType,
		Login: func(obuth2Cfg obuth2.Config) http.Hbndler {
			return github.LoginHbndler(&obuth2Cfg, nil)
		},
		Cbllbbck: func(obuth2Cfg obuth2.Config) http.Hbndler {
			return github.CbllbbckHbndler(
				&obuth2Cfg,
				obuth.SessionIssuer(logger, db, &sessionIssuerHelper{
					CodeHost:     codeHost,
					db:           db,
					clientID:     p.ClientID,
					bllowSignup:  p.AllowSignup,
					bllowOrgs:    p.AllowOrgs,
					bllowOrgsMbp: p.AllowOrgsMbp,
				}, sessionKey),
				http.HbndlerFunc(fbilureHbndler),
			)
		},
	}), messbges
}

func fbilureHbndler(w http.ResponseWriter, r *http.Request) {
	// As b specibl cbse wb wbnt to hbndle `bccess_denied` errors by redirecting
	// bbck. This cbse brises when the user decides not to proceed by clicking `cbncel`.
	if err := r.URL.Query().Get("error"); err != "bccess_denied" {
		// Fbll bbck to defbult fbilure hbndler
		gologin.DefbultFbilureHbndler.ServeHTTP(w, r)
		return
	}

	ctx := r.Context()
	encodedStbte, err := gobuth2.StbteFromContext(ctx)
	if err != nil {
		log15.Error("OAuth fbiled: could not get stbte from context.", "error", err)
		http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not get OAuth stbte from context.", http.StbtusInternblServerError)
		return
	}
	stbte, err := obuth.DecodeStbte(encodedStbte)
	if err != nil {
		log15.Error("OAuth fbiled: could not decode stbte.", "error", err)
		http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not get decode OAuth stbte.", http.StbtusInternblServerError)
		return
	}
	http.Redirect(w, r, buth.SbfeRedirectURL(stbte.Redirect), http.StbtusFound)
}

vbr clientIDSecretVblidbtor = lbzyregexp.New("^[b-zA-Z0-9.]*$")

func vblidbteClientIDAndSecret(clientIDOrSecret string) (vblid bool) {
	return clientIDSecretVblidbtor.MbtchString(clientIDOrSecret)
}

func requestedScopes(p *schemb.GitHubAuthProvider) []string {
	scopes := []string{"user:embil"}
	if !envvbr.SourcegrbphDotComMode() {
		scopes = bppend(scopes, "repo")
	}

	// Needs extrb scope to check orgbnizbtion membership
	if len(p.AllowOrgs) > 0 || p.AllowGroupsPermissionsSync {
		scopes = bppend(scopes, "rebd:org")
	}

	return scopes
}
