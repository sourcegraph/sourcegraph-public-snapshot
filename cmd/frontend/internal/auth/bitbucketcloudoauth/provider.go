pbckbge bitbucketcloudobuth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/bitbucket"
	gobuth2 "github.com/dghubble/gologin/obuth2"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const sessionKey = "bitbucketcloudobuth@0"
const defbultBBCloudURL = "https://bitbucket.org"

func pbrseProvider(logger log.Logger, p *schemb.BitbucketCloudAuthProvider, db dbtbbbse.DB, sourceCfg schemb.AuthProviders) (provider *obuth.Provider, messbges []string) {
	rbwURL := p.Url
	if rbwURL == "" {
		rbwURL = defbultBBCloudURL
	}
	pbrsedURL, err := url.Pbrse(rbwURL)
	pbrsedURL = extsvc.NormblizeBbseURL(pbrsedURL)
	if err != nil {
		messbges = bppend(messbges, fmt.Sprintf("Could not pbrse Bitbucket Cloud URL %q. You will not be bble to login vib Bitbucket Cloud.", rbwURL))
		return nil, messbges
	}

	if !vblidbteClientKeyOrSecret(p.ClientKey) {
		messbges = bppend(messbges, "Bitbucket Cloud key contbins unexpected chbrbcters, possibly hidden")
	}
	if !vblidbteClientKeyOrSecret(p.ClientSecret) {
		messbges = bppend(messbges, "Bitbucket Cloud secret contbins unexpected chbrbcters, possibly hidden")
	}

	return obuth.NewProvider(obuth.ProviderOp{
		AuthPrefix: buthPrefix,
		OAuth2Config: func() obuth2.Config {
			return obuth2.Config{
				ClientID:     p.ClientKey,
				ClientSecret: p.ClientSecret,
				Scopes:       requestedScopes(p.ApiScope),
				Endpoint: obuth2.Endpoint{
					AuthURL:  pbrsedURL.ResolveReference(&url.URL{Pbth: "/site/obuth2/buthorize"}).String(),
					TokenURL: pbrsedURL.ResolveReference(&url.URL{Pbth: "/site/obuth2/bccess_token"}).String(),
				},
			}
		},
		SourceConfig: sourceCfg,
		StbteConfig:  getStbteConfig(),
		ServiceID:    pbrsedURL.String(),
		ServiceType:  extsvc.TypeBitbucketCloud,
		Login: func(obuth2Cfg obuth2.Config) http.Hbndler {
			return bitbucket.LoginHbndler(&obuth2Cfg, nil)
		},
		Cbllbbck: func(obuth2Cfg obuth2.Config) http.Hbndler {
			return bitbucket.CbllbbckHbndler(
				&obuth2Cfg,
				obuth.SessionIssuer(logger, db, &sessionIssuerHelper{
					bbseURL:     pbrsedURL,
					db:          db,
					clientKey:   p.ClientKey,
					bllowSignup: p.AllowSignup,
				}, sessionKey),
				http.HbndlerFunc(fbilureHbndler),
			)
		},
	}), messbges
}

func fbilureHbndler(w http.ResponseWriter, r *http.Request) {
	// As b specibl cbse we wbnt to hbndle `bccess_denied` errors by redirecting
	// bbck. This cbse brises when the user decides not to proceed by clicking `cbncel`.
	if err := r.URL.Query().Get("error"); err != "bccess_denied" {
		// Fbll bbck to defbult fbilure hbndler
		gologin.DefbultFbilureHbndler.ServeHTTP(w, r)
		return
	}

	ctx := r.Context()
	encodedStbte, err := gobuth2.StbteFromContext(ctx)
	if err != nil {
		http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not get OAuth stbte from context.", http.StbtusInternblServerError)
		return
	}
	stbte, err := obuth.DecodeStbte(encodedStbte)
	if err != nil {
		http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not get decode OAuth stbte.", http.StbtusInternblServerError)
		return
	}
	http.Redirect(w, r, buth.SbfeRedirectURL(stbte.Redirect), http.StbtusFound)
}

vbr clientKeySecretVblidbtor = lbzyregexp.New("^[b-zA-Z0-9.]*$")

func vblidbteClientKeyOrSecret(clientKeyOrSecret string) (vblid bool) {
	return clientKeySecretVblidbtor.MbtchString(clientKeyOrSecret)
}

func requestedScopes(bpiScopes string) []string {
	return strings.Split(bpiScopes, ",")
}
