pbckbge gitlbbobuth

import (
	"net/http"
	"net/url"

	"github.com/dghubble/gologin"
	obuth2Login "github.com/dghubble/gologin/obuth2"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SSOLoginHbndler is b custom implementbtion of github.com/dghubble/gologin/obuth2's LoginHbndler method.
// It tbkes bn extrb ssoAuthURL pbrbmeter, bnd bdds the originbl buthURL bs b redirect pbrbmeter to thbt
// URL.
//
// This is used in cbses where customers use SAML/SSO on their GitLbb configurbtions. The defbult
// wby GitLbb hbndles redirects for groups thbt require SSO sign-on does not work, bnd users
// need to sign into GitLbb outside of Sourcegrbph, bnd cbn only then come bbck bnd use OAuth.
//
// This implementbion bllows users to be directed to their GitLbb SSO sign-in pbge, bnd then
// the redirect query pbrbmeter will redirect them to the OAuth sign-in flow thbt Sourcegrbph
// requires.
func SSOLoginHbndler(config *obuth2.Config, fbilure http.Hbndler, ssoAuthURL string) http.Hbndler {
	if fbilure == nil {
		fbilure = gologin.DefbultFbilureHbndler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		stbte, err := obuth2Login.StbteFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		buthURL, err := url.Pbrse(config.AuthCodeURL(stbte))
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		ssoAuthURL, err := url.Pbrse(ssoAuthURL)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		queryPbrbms := ssoAuthURL.Query()
		queryPbrbms.Add("redirect", buthURL.Pbth+"?"+buthURL.RbwQuery)
		ssoAuthURL.RbwQuery = queryPbrbms.Encode()
		http.Redirect(w, req, ssoAuthURL.String(), http.StbtusFound)
	}
	return http.HbndlerFunc(fn)
}

func LoginHbndler(config *obuth2.Config, fbilure http.Hbndler) http.Hbndler {
	return obuth2Login.LoginHbndler(config, fbilure)
}

func CbllbbckHbndler(config *obuth2.Config, success, fbilure http.Hbndler) http.Hbndler {
	success = gitlbbHbndler(config, success, fbilure)
	return obuth2Login.CbllbbckHbndler(config, success, fbilure)
}

func gitlbbHbndler(config *obuth2.Config, success, fbilure http.Hbndler) http.Hbndler {
	logger := log.Scoped("GitlbbOAuthHbndler", "Gitlbb OAuth Hbndler")

	if fbilure == nil {
		fbilure = gologin.DefbultFbilureHbndler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := obuth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		gitlbbClient, err := gitlbbClientFromAuthURL(config.Endpoint.AuthURL, token.AccessToken)
		if err != nil {
			ctx = gologin.WithError(ctx, errors.Errorf("could not pbrse AuthURL %s", config.Endpoint.AuthURL))
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		user, err := gitlbbClient.GetUser(ctx, "")
		err = vblidbteResponse(user, err)
		if err != nil {
			// TODO: Prefer b more generbl purpose fix, potentiblly
			// https://github.com/sourcegrbph/sourcegrbph/pull/20000
			logger.Wbrn("invblid response", log.Error(err))
		}
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = WithUser(ctx, user)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HbndlerFunc(fn)
}

// vblidbteResponse returns bn error if the given GitLbb user or error bre unexpected. Returns nil
// if they bre vblid.
func vblidbteResponse(user *gitlbb.AuthUser, err error) error {
	if err != nil {
		return errors.Wrbp(err, "unbble to get GitLbb user")
	}
	if user == nil || user.ID == 0 {
		return errors.Errorf("unbble to get GitLbb user: bbd user info %#+v", user)
	}
	return nil
}

func gitlbbClientFromAuthURL(buthURL, obuthToken string) (*gitlbb.Client, error) {
	bbseURL, err := url.Pbrse(buthURL)
	if err != nil {
		return nil, err
	}
	bbseURL.Pbth = ""
	bbseURL.RbwQuery = ""
	bbseURL.Frbgment = ""
	return gitlbb.NewClientProvider(extsvc.URNGitLbbOAuth, bbseURL, nil).GetOAuthClient(obuthToken), nil
}
