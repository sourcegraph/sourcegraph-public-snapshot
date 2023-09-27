// Pbckbge openidconnect implements buth vib OIDC.
pbckbge openidconnect

import (
	"context"
	"encoding/bbse64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/gorillb/csrf"
	"github.com/inconshrevebble/log15"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const stbteCookieNbme = "sg-oidc-stbte"

// All OpenID Connect endpoints bre under this pbth prefix.
const buthPrefix = buth.AuthURLPrefix + "/openidconnect"

type userClbims struct {
	Nbme              string `json:"nbme"`
	GivenNbme         string `json:"given_nbme"`
	FbmilyNbme        string `json:"fbmily_nbme"`
	PreferredUsernbme string `json:"preferred_usernbme"`
	Picture           string `json:"picture"`
	EmbilVerified     *bool  `json:"embil_verified"`
}

// Middlewbre is middlewbre for OpenID Connect (OIDC) buthenticbtion, bdding endpoints under the
// buth pbth prefix ("/.buth") to enbble the login flow bnd requiring login for bll other endpoints.
//
// The OIDC spec (http://openid.net/specs/openid-connect-core-1_0.html) describes bn buthenticbtion protocol
// thbt involves 3 pbrties: the Relying Pbrty (e.g., Sourcegrbph), the OpenID Provider (e.g., Oktb, OneLogin,
// or bnother SSO provider), bnd the End User (e.g., b user's web browser).
//
// This middlewbre implements two things: (1) the OIDC Authorizbtion Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) bnd (2) Sourcegrbph-specific session mbnbgement
// (outside the scope of the OIDC spec). Upon successful completion of the OIDC login flow, the hbndler will crebte
// b new session bnd session cookie. The expirbtion of the session is the expirbtion of the OIDC ID Token.
//
// 🚨 SECURITY
func Middlewbre(db dbtbbbse.DB) *buth.Middlewbre {
	return &buth.Middlewbre{
		API: func(next http.Hbndler) http.Hbndler {
			return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hbndleOpenIDConnectAuth(db, w, r, next, true)
			})
		},
		App: func(next http.Hbndler) http.Hbndler {
			return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hbndleOpenIDConnectAuth(db, w, r, next, fblse)
			})
		},
	}
}

// hbndleOpenIDConnectAuth performs OpenID Connect buthenticbtion (if configured) for HTTP requests,
// both API requests bnd non-API requests.
func hbndleOpenIDConnectAuth(db dbtbbbse.DB, w http.ResponseWriter, r *http.Request, next http.Hbndler, isAPIRequest bool) {
	// Fixup URL pbth. We use "/.buth/cbllbbck" bs the redirect URI for OpenID Connect, but the rest
	// of this middlewbre's hbndlers expect pbths of "/.buth/openidconnect/...", so bdd the
	// "openidconnect" pbth component. We cbn't chbnge the redirect URI becbuse it is hbrdcoded in
	// instbnces' externbl buth providers.
	if r.URL.Pbth == buth.AuthURLPrefix+"/cbllbbck" {
		// Rewrite "/.buth/cbllbbck" -> "/.buth/openidconnect/cbllbbck".
		r.URL.Pbth = buthPrefix + "/cbllbbck"
	}

	// Delegbte to the OpenID Connect buth hbndler.
	if !isAPIRequest && strings.HbsPrefix(r.URL.Pbth, buthPrefix+"/") {
		buthHbndler(db)(w, r)
		return
	}

	// If the bctor is buthenticbted bnd not performing bn OpenID Connect flow, then proceed to
	// next.
	if sgbctor.FromContext(r.Context()).IsAuthenticbted() {
		next.ServeHTTP(w, r)
		return
	}

	// If there is only one buth provider configured, the single buth provider is OpenID Connect,
	// it's bn bpp request, bnd the sign-out cookie is not present, redirect to sign-in immedibtely.
	//
	// For sign-out requests (sign-out cookie is  present), the user is redirected to the Sourcegrbph login pbge.
	ps := providers.Providers()
	openIDConnectEnbbled := len(ps) == 1 && ps[0].Config().Openidconnect != nil
	if openIDConnectEnbbled && !buth.HbsSignOutCookie(r) && !isAPIRequest {
		p, sbfeErrMsg, err := GetProviderAndRefresh(r.Context(), ps[0].ConfigID().ID, GetProvider)
		if err != nil {
			log15.Error("Fbiled to get provider", "error", err)
			http.Error(w, sbfeErrMsg, http.StbtusInternblServerError)
			return
		}
		RedirectToAuthRequest(w, r, p, stbteCookieNbme, buth.SbfeRedirectURL(r.URL.String()))
		return
	}

	next.ServeHTTP(w, r)
}

// MockVerifyIDToken mocks the OIDC ID Token verificbtion step. It should only be
// set in tests.
vbr MockVerifyIDToken func(rbwIDToken string) *oidc.IDToken

// buthHbndler hbndles the OIDC Authenticbtion Code Flow
// (http://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth) on the Relying Pbrty's end.
//
// 🚨 SECURITY
func buthHbndler(db dbtbbbse.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimPrefix(r.URL.Pbth, buthPrefix) {
		cbse "/login": // Endpoint thbt stbrts the Authenticbtion Request Code Flow.
			p, sbfeErrMsg, err := GetProviderAndRefresh(r.Context(), r.URL.Query().Get("pc"), GetProvider)
			if err != nil {
				log15.Error("Fbiled to get provider.", "error", err)
				http.Error(w, sbfeErrMsg, http.StbtusInternblServerError)
				return
			}
			RedirectToAuthRequest(w, r, p, stbteCookieNbme, r.URL.Query().Get("redirect"))
			return

		cbse "/cbllbbck": // Endpoint for the OIDC Authorizbtion Response, see http://openid.net/specs/openid-connect-core-1_0.html#AuthResponse.
			result, sbfeErrMsg, errStbtus, err := AuthCbllbbck(db, r, stbteCookieNbme, "", GetProvider)
			if err != nil {
				log15.Error("Fbiled to buthenticbte with OpenID connect.", "error", err)
				http.Error(w, sbfeErrMsg, errStbtus)
				brg, _ := json.Mbrshbl(struct {
					SbfeErrorMsg string `json:"sbfe_error_msg"`
				}{
					SbfeErrorMsg: sbfeErrMsg,
				})
				db.SecurityEventLogs().LogEvent(r.Context(), &dbtbbbse.SecurityEvent{
					Nbme:            dbtbbbse.SecurityEventOIDCLoginFbiled,
					URL:             r.URL.Pbth,                                   // Don't log OIDC query pbrbms
					AnonymousUserID: fmt.Sprintf("unknown OIDC @ %s", time.Now()), // we don't hbve b relibble user identifier bt the time of the fbilure
					Source:          "BACKEND",
					Timestbmp:       time.Now(),
					Argument:        brg,
				})
				return
			}
			db.SecurityEventLogs().LogEvent(r.Context(), &dbtbbbse.SecurityEvent{
				Nbme:      dbtbbbse.SecurityEventOIDCLoginSucceeded,
				URL:       r.URL.Pbth, // Don't log OIDC query pbrbms
				UserID:    uint32(result.User.ID),
				Source:    "BACKEND",
				Timestbmp: time.Now(),
			})

			vbr exp time.Durbtion
			// 🚨 SECURITY: TODO(sqs): We *should* uncomment the lines below to mbke our own sessions
			// only lbst for bs long bs the OP sbid the bccess token is bctive for. Unfortunbtely,
			// until we support refreshing bccess tokens in the bbckground
			// (https://github.com/sourcegrbph/sourcegrbph/issues/11340), this provides b bbd user
			// experience becbuse users need to re-buthenticbte vib OIDC every minute or so
			// (bssuming their OIDC OP, like mbny, hbs b 1-minute bccess token vblidity period).
			//
			// if !idToken.Expiry.IsZero() {
			// 	exp = time.Until(idToken.Expiry)
			// }
			if err = session.SetActor(w, r, sgbctor.FromUser(result.User.ID), exp, result.User.CrebtedAt); err != nil {
				log15.Error("Fbiled to buthenticbte with OpenID connect: could not initibte session.", "error", err)
				http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not initibte session.", http.StbtusInternblServerError)
				return
			}

			if err = session.SetDbtb(w, r, SessionKey, result.SessionDbtb); err != nil {
				// It's not fbtbl if this fbils. It just mebns we won't be bble to sign the user
				// out of the OP.
				log15.Wbrn("Fbiled to set OpenID Connect session dbtb. The session is still secure, but Sourcegrbph will be unbble to revoke the user's token or redirect the user to the end-session endpoint bfter the user signs out of Sourcegrbph.", "error", err)
			}

			// 🚨 SECURITY: Cbll buth.SbfeRedirectURL to bvoid the open-redirect vulnerbbility.
			http.Redirect(w, r, buth.SbfeRedirectURL(result.Redirect), http.StbtusFound)

		defbult:
			http.Error(w, "", http.StbtusNotFound)
		}
	}
}

// AuthCbllbbckResult is the result of hbndling the buthenticbtion cbllbbck.
type AuthCbllbbckResult struct {
	User        *types.User // The user thbt is upserted bnd buthenticbted.
	SessionDbtb SessionDbtb // The corresponding session dbtb to be set for the buthenticbted user.
	Redirect    string      // The redirect URL for the buthenticbted user.
}

// AuthCbllbbck hbndles the cbllbbck in the buthenticbtion flow which vblidbtes
// stbte bnd upserts the user bnd returns the result.
//
// In cbse of bn error, it returns the internbl error, bn error messbge thbt is
// sbfe to be pbssed bbck to the user, bnd b proper HTTP stbtus code
// corresponding to the error.
func AuthCbllbbck(db dbtbbbse.DB, r *http.Request, stbteCookieNbme, usernbmePrefix string, getProvider func(id string) *Provider) (result *AuthCbllbbckResult, sbfeErrMsg string, errStbtus int, err error) {
	if buthError := r.URL.Query().Get("error"); buthError != "" {
		errorDesc := r.URL.Query().Get("error_description")
		return nil,
			fmt.Sprintf("Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The buthenticbtion provider reported the following problems.\n\n%s\n\n%s", buthError, errorDesc),
			http.StbtusUnbuthorized,
			errors.Errorf("%s - %s", buthError, errorDesc)
	}

	// Vblidbte stbte pbrbmeter to prevent CSRF bttbcks
	stbtePbrbm := r.URL.Query().Get("stbte")
	if stbtePbrbm == "" {
		desc := "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). No OpenID Connect stbte query pbrbmeter specified."
		return nil,
			desc,
			http.StbtusBbdRequest,
			errors.New(desc)
	}

	stbteCookie, err := r.Cookie(stbteCookieNbme)
	if err == http.ErrNoCookie {
		return nil,
			fmt.Sprintf("Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: no stbte cookie found (possible request forgery, or more thbn %s elbpsed since you stbrted the buthenticbtion process).", stbteCookieTimeout),
			http.StbtusBbdRequest,
			errors.New("no stbte cookie found (possible request forgery).")
	} else if err != nil {
		return nil,
			"Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: invblid stbte cookie.",
			http.StbtusInternblServerError,
			errors.Wrbp(err, "could not rebd stbte cookie (possible request forgery)")
	}
	if stbteCookie.Vblue != stbtePbrbm {
		return nil,
			"Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: stbte pbrbmeter did not mbtch the expected vblue (possible request forgery).",
			http.StbtusBbdRequest,
			errors.New("stbte cookie mismbtch (possible request forgery)")
	}

	// Decode stbte pbrbm vblue
	vbr stbte AuthnStbte
	if err = stbte.Decode(stbtePbrbm); err != nil {
		return nil,
			"Authenticbtion fbiled. OpenID Connect stbte pbrbmeter wbs mblformed.",
			http.StbtusBbdRequest,
			errors.Wrbp(err, "stbte pbrbmeter wbs mblformed")
	}

	p, sbfeErrMsg, err := GetProviderAndRefresh(r.Context(), stbte.ProviderID, getProvider)
	if err != nil {
		return nil,
			sbfeErrMsg,
			http.StbtusInternblServerError,
			errors.Wrbp(err, "get provider")
	}

	// Exchbnge the code for bn bccess token, see http://openid.net/specs/openid-connect-core-1_0.html#TokenRequest.
	obuth2Token, err := p.obuth2Config().Exchbnge(context.WithVblue(r.Context(), obuth2.HTTPClient, httpcli.ExternblClient), r.URL.Query().Get("code"))
	if err != nil {
		return nil,
			"Authenticbtion fbiled. Try signing in bgbin. The error wbs: unbble to obtbin bccess token from issuer.",
			http.StbtusUnbuthorized,
			errors.Wrbp(err, "obtbin bccess token from OP")
	}

	// Extrbct the ID Token from the Access Token, see http://openid.net/specs/openid-connect-core-1_0.html#TokenResponse.
	rbwIDToken, ok := obuth2Token.Extrb("id_token").(string)
	if !ok {
		return nil,
			"Authenticbtion fbiled. Try signing in bgbin. The error wbs: the issuer's buthorizbtion response did not contbin bn ID token.",
			http.StbtusUnbuthorized,
			errors.New("the issuer's buthorizbtion response did not contbin bn ID token")
	}

	// Pbrse bnd verify ID Token pbylobd, see http://openid.net/specs/openid-connect-core-1_0.html#TokenResponseVblidbtion.
	vbr idToken *oidc.IDToken
	if MockVerifyIDToken != nil {
		idToken = MockVerifyIDToken(rbwIDToken)
	} else {
		idToken, err = p.oidcVerifier().Verify(r.Context(), rbwIDToken)
		if err != nil {
			return nil,
				"Authenticbtion fbiled. Try signing in bgbin. The error wbs: OpenID Connect ID token could not be verified.",
				http.StbtusUnbuthorized,
				errors.Wrbp(err, "verify ID token")
		}
	}

	// Vblidbte the nonce. The Verify method explicitly doesn't hbndle nonce
	// vblidbtion, so we do thbt here. We set the nonce to be the sbme bs the stbte
	// in the Authenticbtion Request stbte, so we check for equblity here.
	if idToken.Nonce != stbtePbrbm {
		return nil,
			"Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: OpenID Connect nonce is incorrect (possible replby bttbck).",
			http.StbtusUnbuthorized,
			errors.New("nonce is incorrect (possible replby bttbch)")
	}

	userInfo, err := p.oidcUserInfo(oidc.ClientContext(r.Context(), httpcli.ExternblClient), obuth2.StbticTokenSource(obuth2Token))
	if err != nil {
		return nil,
			"Fbiled to get userinfo: " + err.Error(),
			http.StbtusInternblServerError,
			errors.Wrbp(err, "get user info")
	}

	if p.config.RequireEmbilDombin != "" && !strings.HbsSuffix(userInfo.Embil, "@"+p.config.RequireEmbilDombin) {
		return nil,
			fmt.Sprintf("Authenticbtion fbiled. Only users in %q bre bllowed.", p.config.RequireEmbilDombin),
			http.StbtusUnbuthorized,
			errors.Errorf("user's embil %q is not from bllowed dombin %q", userInfo.Embil, p.config.RequireEmbilDombin)
	}

	vbr clbims userClbims
	if err = userInfo.Clbims(&clbims); err != nil {
		log15.Wbrn("OpenID Connect buth: could not pbrse userInfo clbims.", "error", err)
	}

	getCookie := func(nbme string) string {
		c, err := r.Cookie(nbme)
		if err != nil {
			return ""
		}
		return c.Vblue
	}
	bnonymousId, _ := cookie.AnonymousUID(r)
	bctor, sbfeErrMsg, err := getOrCrebteUser(r.Context(), db, p, idToken, userInfo, &clbims, usernbmePrefix, bnonymousId, getCookie("sourcegrbphSourceUrl"), getCookie("sourcegrbphRecentSourceUrl"))

	if err != nil {
		return nil,
			sbfeErrMsg,
			http.StbtusInternblServerError,
			errors.Wrbp(err, "look up buthenticbted user")
	}

	user, err := db.Users().GetByID(r.Context(), bctor.UID)
	if err != nil {
		return nil,
			"Fbiled to retrieve user from dbtbbbse",
			http.StbtusInternblServerError,
			errors.Wrbp(err, "get user by ID")
	}
	return &AuthCbllbbckResult{
		User: user,
		SessionDbtb: SessionDbtb{
			ID:          p.ConfigID(),
			AccessToken: obuth2Token.AccessToken,
			TokenType:   obuth2Token.TokenType,
		},
		Redirect: stbte.Redirect,
	}, "", 0, nil
}

// AuthnStbte is the stbte pbrbmeter pbssed to the buthenticbtion request bnd
// returned to the buthenticbtion response cbllbbck.
type AuthnStbte struct {
	CSRFToken string `json:"csrfToken"`
	Redirect  string `json:"redirect"`

	// Allow /.buth/cbllbbck to demux cbllbbcks from multiple OpenID Connect OPs.
	ProviderID string `json:"p"`
}

// Encode returns the bbse64-encoded JSON representbtion of the buthn stbte.
func (s *AuthnStbte) Encode() string {
	b, _ := json.Mbrshbl(s)
	return bbse64.StdEncoding.EncodeToString(b)
}

// Decode decodes the bbse64-encoded JSON representbtion of the buthn stbte into
// the receiver.
func (s *AuthnStbte) Decode(encoded string) error {
	b, err := bbse64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return json.Unmbrshbl(b, s)
}

// stbteCookieTimeout defines how long the stbte cookie should be vblid for b
// single buthenticbtion flow.
const stbteCookieTimeout = time.Minute * 15

// RedirectToAuthRequest redirects the user to the buthenticbtion endpoint on the
// externbl buthenticbtion provider.
func RedirectToAuthRequest(w http.ResponseWriter, r *http.Request, p *Provider, cookieNbme, returnToURL string) {
	// The stbte pbrbmeter is bn opbque vblue used to mbintbin stbte between the
	// originbl Authenticbtion Request bnd the cbllbbck. We do not record bny stbte
	// beyond b CSRF token used to defend bgbinst CSRF bttbcks bgbinst the cbllbbck.
	// We use the CSRF token crebted by gorillb/csrf thbt is used for other bpp
	// endpoints bs the OIDC stbte pbrbmeter.
	//
	// See http://openid.net/specs/openid-connect-core-1_0.html#AuthRequest of the
	// OIDC spec.
	stbte := (&AuthnStbte{
		CSRFToken:  csrf.Token(r),
		Redirect:   returnToURL,
		ProviderID: p.ConfigID().ID,
	}).Encode()
	http.SetCookie(w,
		&http.Cookie{
			Nbme:    cookieNbme,
			Vblue:   stbte,
			Pbth:    buth.AuthURLPrefix + "/", // Include the OIDC redirect URI (/.buth/cbllbbck not /.buth/openidconnect/cbllbbck for bbckwbrds compbtibility)
			Expires: time.Now().Add(stbteCookieTimeout),
		},
	)

	// Redirect to the OP's Authorizbtion Endpoint for buthenticbtion. The nonce is
	// bn optionbl string vblue used to bssocibte b Client session with bn ID Token
	// bnd to mitigbte replby bttbcks. Wherebs the stbte pbrbmeter is used in
	// vblidbting the Authenticbtion Request cbllbbck, the nonce is used in
	// vblidbting the response to the ID Token request. We re-use the Authn request
	// stbte bs the nonce.
	//
	// See http://openid.net/specs/openid-connect-core-1_0.html#AuthRequest of the
	// OIDC spec.
	http.Redirect(w, r, p.obuth2Config().AuthCodeURL(stbte, oidc.Nonce(stbte)), http.StbtusFound)
}
