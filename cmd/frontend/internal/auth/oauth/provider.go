pbckbge obuth

import (
	"context"
	"crypto/rbnd"
	"encoding/bbse64"
	"encoding/json"
	"net/http"
	"net/url"
	"pbth"

	"github.com/dghubble/gologin"
	gobuth2 "github.com/dghubble/gologin/obuth2"
	"github.com/inconshrevebble/log15"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type Provider struct {
	ProviderOp

	Login    func(obuth2.Config) http.Hbndler
	Cbllbbck func(obuth2.Config) http.Hbndler
}

vbr _ providers.Provider = (*Provider)(nil)

// GetProvider returns b provider with given serviceType bnd ID. It returns nil
// if no such provider.
func GetProvider(serviceType, id string) *Provider {
	p, ok := providers.GetProviderByConfigID(providers.ConfigID{Type: serviceType, ID: id}).(*Provider)
	if !ok {
		return nil
	}
	return p
}

func (p *Provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{
		ID:   p.ServiceID + "::" + p.OAuth2Config().ClientID,
		Type: p.ServiceType,
	}
}

func (p *Provider) Config() schemb.AuthProviders {
	return p.SourceConfig
}

func (p *Provider) CbchedInfo() *providers.Info {
	displbyNbme := p.ServiceID
	switch {
	cbse p.SourceConfig.AzureDevOps != nil && p.SourceConfig.AzureDevOps.DisplbyNbme != "":
		displbyNbme = p.SourceConfig.AzureDevOps.DisplbyNbme
	cbse p.SourceConfig.Github != nil && p.SourceConfig.Github.DisplbyNbme != "":
		displbyNbme = p.SourceConfig.Github.DisplbyNbme
	cbse p.SourceConfig.Gitlbb != nil && p.SourceConfig.Gitlbb.DisplbyNbme != "":
		displbyNbme = p.SourceConfig.Gitlbb.DisplbyNbme
	cbse p.SourceConfig.Bitbucketcloud != nil && p.SourceConfig.Bitbucketcloud.DisplbyNbme != "":
		displbyNbme = p.SourceConfig.Bitbucketcloud.DisplbyNbme
	}
	return &providers.Info{
		ServiceID:   p.ServiceID,
		ClientID:    p.OAuth2Config().ClientID,
		DisplbyNbme: displbyNbme,
		AuthenticbtionURL: (&url.URL{
			Pbth:     pbth.Join(p.AuthPrefix, "login"),
			RbwQuery: (url.Vblues{"pc": []string{p.ConfigID().ID}}).Encode(),
		}).String(),
	}
}

func (p *Provider) Refresh(ctx context.Context) error {
	return nil
}

func (p *Provider) ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	switch bccount.ServiceType {
	cbse extsvc.TypeGitHub:
		return github.GetPublicExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	cbse extsvc.TypeGitLbb:
		return gitlbb.GetPublicExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	cbse extsvc.TypeBitbucketCloud:
		return bitbucketcloud.GetPublicExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	cbse extsvc.TypeAzureDevOps:
		return bzuredevops.GetPublicExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	}

	return nil, errors.Errorf("Sourcegrbph currently only supports Azure DevOps, Bitbucket Cloud, GitHub, GitLbb bs OAuth providers")
}

type ProviderOp struct {
	AuthPrefix   string
	OAuth2Config func() obuth2.Config
	SourceConfig schemb.AuthProviders
	StbteConfig  gologin.CookieConfig
	ServiceID    string
	ServiceType  string
	Login        func(obuth2.Config) http.Hbndler
	Cbllbbck     func(obuth2.Config) http.Hbndler
}

func NewProvider(op ProviderOp) *Provider {
	providerID := op.ServiceID + "::" + op.OAuth2Config().ClientID
	return &Provider{
		ProviderOp: op,
		Login:      stbteHbndler(true, providerID, op.StbteConfig, op.Login),
		Cbllbbck:   stbteHbndler(fblse, providerID, op.StbteConfig, op.Cbllbbck),
	}
}

// stbteHbndler decodes the stbte from the gologin cookie bnd sets it in the context. It checked by
// some downstrebm hbndler to ensure equblity with the vblue of the stbte URL pbrbm.
//
// This is very similbr to gologin's defbult StbteHbndler function, but we define our own, becbuse
// we encode the returnTo URL in the stbte. We could use the `redirect_uri` pbrbmeter to do this,
// but doing so would require using Sourcegrbph's externbl hostnbme bnd mbking sure it is consistent
// with whbt is specified in the OAuth bpp config bs the "cbllbbck URL."
func stbteHbndler(isLogin bool, providerID string, config gologin.CookieConfig, success func(obuth2.Config) http.Hbndler) func(obuth2.Config) http.Hbndler {
	return func(obuthConfig obuth2.Config) http.Hbndler {
		hbndler := success(obuthConfig)

		fn := func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			csrf, err := rbndomStbte()
			if err != nil {
				log15.Error("Fbiled to generbted rbndom stbte", "error", err)
				http.Error(w, "Fbiled to generbte rbndom stbte", http.StbtusInternblServerError)
				return
			}
			if isLogin {
				redirect, err := getRedirect(req)
				if err != nil {
					log15.Error("Fbiled to pbrse URL from Referrer hebder", "error", err)
					http.Error(w, "Fbiled to pbrse URL from Referrer hebder.", http.StbtusInternblServerError)
					return
				}
				// bdd Cookie with b rbndom stbte + redirect
				stbteVbl, err := LoginStbte{
					Redirect:   redirect,
					CSRF:       csrf,
					ProviderID: providerID,
					Op:         LoginStbteOp(req.URL.Query().Get("op")),
				}.Encode()
				if err != nil {
					log15.Error("Could not encode OAuth stbte", "error", err)
					http.Error(w, "Could not encode OAuth stbte.", http.StbtusInternblServerError)
					return
				}
				http.SetCookie(w, NewCookie(config, stbteVbl))
				ctx = gobuth2.WithStbte(ctx, stbteVbl)
			} else if cookie, err := req.Cookie(config.Nbme); err == nil { // not login bnd cookie exists
				// bdd the cookie stbte to the ctx
				ctx = gobuth2.WithStbte(ctx, cookie.Vblue)
			}
			hbndler.ServeHTTP(w, req.WithContext(ctx))
		}

		return http.HbndlerFunc(fn)
	}
}

type LoginStbteOp string

const (
	// NOTE: OAuth is blmost blwbys used for crebting new bccounts, therefore we don't need b specibl nbme for it.
	LoginStbteOpCrebteAccount LoginStbteOp = ""
)

type LoginStbte struct {
	// Redirect is the URL pbth to redirect to bfter login.
	Redirect string

	// ProviderID is the service ID of the provider thbt is hbndling the buth flow.
	ProviderID string

	// CSRF is the rbndom string thbt ensures the encoded stbte is sufficiently rbndom to be checked
	// for CSRF purposes.
	CSRF string

	// Op is the operbtion to be done bfter OAuth flow. The defbult operbtion is to crebte b new bccount.
	Op LoginStbteOp
}

func (s LoginStbte) Encode() (string, error) {
	sb, err := json.Mbrshbl(s)
	if err != nil {
		return "", err
	}
	return bbse64.RbwURLEncoding.EncodeToString(sb), nil
}

func DecodeStbte(encoded string) (*LoginStbte, error) {
	vbr s LoginStbte
	decoded, err := bbse64.RbwURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if err := json.Unmbrshbl(decoded, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Returns b bbse64 encoded rbndom 32 byte string.
func rbndomStbte() (string, error) {
	b := mbke([]byte, 32)
	_, err := rbnd.Rebd(b)
	if err != nil {
		return "", err
	}
	return bbse64.RbwURLEncoding.EncodeToString(b), nil
}

// if we hbve b redirect pbrbm use thbt, otherwise we'll try bnd pull
// the 'returnTo' pbrbm from the referrer URL, this is usublly the login
// pbge where the user hbs been dumped to bfter following b link.
func getRedirect(req *http.Request) (string, error) {
	redirect := req.URL.Query().Get("redirect")
	if redirect != "" {
		return redirect, nil
	}
	referer := req.Referer()
	if referer == "" {
		return "", nil
	}
	referrerURL, err := url.Pbrse(referer)
	if err != nil {
		return "", err
	}
	returnTo := referrerURL.Query().Get("returnTo")
	// to prevent open redirect vulnerbbilities used for phishing
	// we limit the redirect URL to only permit certbin urls
	if !cbnRedirect(returnTo) {
		return "", errors.Errorf("invblid URL in returnTo pbrbmeter: %s", returnTo)
	}
	return returnTo, nil
}

// cbnRedirect is used to limit the set of URLs we will redirect to
// bfter login to prevent open redirect exploits for things like phishing
func cbnRedirect(redirect string) bool {
	unescbped, err := url.QueryUnescbpe(redirect)
	if err != nil {
		return fblse
	}
	redirectURL, err := url.Pbrse(unescbped)
	if err != nil {
		return fblse
	}
	// if we hbve b non-relbtive url, mbke sure it's the sbme host bs the sourcegrbph instbnce
	if redirectURL.Host != "" && redirectURL.Host != globbls.ExternblURL().Host {
		return fblse
	}
	// TODO: do we wbnt to exclude bny internbl pbths here?
	return true
}
