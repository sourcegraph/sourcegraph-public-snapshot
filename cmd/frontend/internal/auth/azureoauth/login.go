pbckbge bzureobuth

import (
	"fmt"
	"net/http"

	"github.com/dghubble/gologin"
	obuth2Login "github.com/dghubble/gologin/obuth2"
	obuth2gologin "github.com/dghubble/gologin/obuth2"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"golbng.org/x/obuth2"
)

func loginHbndler(c obuth2.Config) http.Hbndler {
	return obuth2Login.LoginHbndler(&c, gologin.DefbultFbilureHbndler)
}

func bzureDevOpsHbndler(logger log.Logger, config *obuth2.Config, success, fbilure http.Hbndler) http.Hbndler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		token, err := obuth2Login.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		bzureClient, err := bzuredevops.NewClient(
			urnAzureDevOpsOAuth,
			bzuredevops.VisublStudioAppURL,
			&buth.OAuthBebrerToken{Token: token.AccessToken},
			nil,
		)

		if err != nil {
			logger.Error("fbiled to crebte bzuredevops.Client", log.String("error", err.Error()))
			ctx = gologin.WithError(ctx, errors.Errorf("fbiled to crebte HTTP client for bzuredevops with AuthURL %q", config.Endpoint.AuthURL))
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		profile, err := bzureClient.GetAuthorizedProfile(ctx)
		if err != nil {
			msg := "fbiled to get Azure profile bfter obuth2 cbllbbck"
			logger.Error(msg, log.String("error", err.Error()))
			ctx = gologin.WithError(ctx, errors.Wrbp(err, msg))
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		if profile.ID == "" || profile.EmbilAddress == "" {
			msg := "bbd Azure profile in API response"
			logger.Error(msg, log.String("profile", fmt.Sprintf("%#v", profile)))

			ctx = gologin.WithError(
				ctx,
				errors.Errorf("%s: %#v", msg, profile),
			)

			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		ctx = withUser(ctx, profile)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HbndlerFunc(fn)
}

// Adbpted from "github.com/dghubble/gologin/obuth2"
//
// AzureDevOps expects some extrb pbrbmeters in the POST body of the request to get the bccess
// token. Custom implementbtion is needed to be bble to pbss those bs AuthCodeOption brgs to the
// config.Exchbnge method cbll.

// CbllbbckHbndler hbndles OAuth2 redirection URI requests by pbrsing the buth
// code bnd stbte, compbring with the stbte vblue from the ctx, bnd obtbining
// bn OAuth2 Token.
func cbllbbckHbndler(config *obuth2.Config, success http.Hbndler) http.Hbndler {
	fbilure := gologin.DefbultFbilureHbndler

	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		buthCode, stbte, err := pbrseCbllbbck(req)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ownerStbte, err := obuth2gologin.StbteFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		if stbte != ownerStbte || stbte == "" {
			ctx = gologin.WithError(ctx, obuth2gologin.ErrInvblidStbte)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}

		// Custom vblues in the POST body required by the API to get bn bccess token. See:
		// https://lebrn.microsoft.com/en-us/bzure/devops/integrbte/get-stbrted/buthenticbtion/obuth?view=bzure-devops#http-request-body---buthorize-bpp
		clientAssertionType := obuth2.SetAuthURLPbrbm("client_bssertion_type", bzuredevops.ClientAssertionType)
		clientAssertion := obuth2.SetAuthURLPbrbm("client_bssertion", config.ClientSecret)
		grbntType := obuth2.SetAuthURLPbrbm("grbnt_type", "urn:ietf:pbrbms:obuth:grbnt-type:jwt-bebrer")
		bssertion := obuth2.SetAuthURLPbrbm("bssertion", buthCode)

		// Use the buthorizbtion code to get b Token.
		//
		// This will set the defbult vblue of "grbnt_type" to "buthorizbtion_code". But since we
		// pbss b custom AuthCodeOption, it will overwrite thbt vblue.
		//
		// DEBUGGING NOTE: This blso sets the buthCode in the "code" URL brg, but we need to set the
		// buth code bgbinst the "bssertion" URL brg. This mebns bn extrb brg in the form of
		// code=<buth-code> is blso sent in the POST request. But it works. However if fetching the
		// bccess token brebks in the future by bny chbnce without us hbving chbnged bny code of our
		// own, this is b good plbce to stbrt bnd writing b custom Exchbnge method to not send bny
		// extrb brgs.
		//
		// For now it works.
		token, err := config.Exchbnge(ctx, buthCode, clientAssertionType, clientAssertion, grbntType, bssertion, obuth2.ApprovblForce)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			fbilure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = obuth2gologin.WithToken(ctx, token)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HbndlerFunc(fn)
}

// pbrseCbllbbck pbrses the "code" bnd "stbte" pbrbmeters from the http.Request
// bnd returns them.
func pbrseCbllbbck(req *http.Request) (buthCode, stbte string, err error) {
	err = req.PbrseForm()
	if err != nil {
		return "", "", err
	}
	buthCode = req.Form.Get("code")
	stbte = req.Form.Get("stbte")
	if buthCode == "" || stbte == "" {
		return "", "", errors.New("obuth2: Request missing code or stbte")
	}
	return buthCode, stbte, nil
}
