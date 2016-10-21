package oauth2client

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/canonicalurl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/google.golang.org/api/source/v1"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
)

var googleNonceCookiePath = router.Rel.URLTo(router.GoogleOAuth2Receive).Path

func auth0GoogleConfigWithRedirectURL() *oauth2.Config {
	config := *auth.Auth0Config
	config.RedirectURL = conf.AppURL.ResolveReference(router.Rel.URLTo(router.GoogleOAuth2Receive)).String()
	return &config
}

// ServeGoogleOAuth2Initiate generates the OAuth2 authorize URL
// (including a nonce state value, also stored in a cookie) and
// redirects the client to that URL.
func ServeGoogleOAuth2Initiate(w http.ResponseWriter, r *http.Request) error {
	returnTo, err := returnto.URLFromRequest(r)
	if err != nil {
		log15.Warn("Invalid return-to URL provided to OAuth2 flow initiation; ignoring.", "err", err)
	}

	// Remove UTM campaign params to avoid double
	// attribution. TODO(sqs): consider doing this on the frontend in
	// JS so we centralize usage analytics there.
	returnTo = canonicalurl.FromURL(returnTo)

	scopes := []string{
		googleoauth2.UserinfoProfileScope,
		googleoauth2.UserinfoEmailScope,

		source.CloudPlatformScope, // For source.projects.repos.list method.
	}

	return googleOAuth2Initiate(w, r, scopes, returnTo.String())
}

func googleOAuth2Initiate(w http.ResponseWriter, r *http.Request, scopes []string, returnTo string) error {
	nonce := randstring.NewLen(32)
	http.SetCookie(w, &http.Cookie{
		Name:    "nonce",
		Value:   nonce,
		Path:    "",
		Expires: time.Now().Add(10 * time.Minute),
	})

	http.Redirect(w, r, auth0GoogleConfigWithRedirectURL().AuthCodeURL(nonce+":"+returnTo,
		oauth2.SetAuthURLParam("connection", "google-oauth2"),
		oauth2.SetAuthURLParam("connection_scope", strings.Join(scopes, ",")),
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce, // Might be mandatory; refresh token isn't populated otherwise.
	), http.StatusSeeOther)
	return nil
}

func ServeGoogleOAuth2Receive(w http.ResponseWriter, r *http.Request) (err error) {
	parts := strings.SplitN(r.URL.Query().Get("state"), ":", 2)
	if len(parts) != 2 {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("invalid OAuth2 authorize client state")}
	}

	nonceCookie, err := r.Cookie("nonce")
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "nonce",
		Path:   "/",
		MaxAge: -1,
	})
	if len(parts) != 2 || nonceCookie.Value == "" || parts[0] != nonceCookie.Value {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("invalid state")}
	}
	returnTo := parts[1]

	code := r.URL.Query().Get("code")
	token, err := auth0GoogleConfigWithRedirectURL().Exchange(r.Context(), code)
	if err != nil {
		return err
	}
	if !token.Valid() {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("exchanging auth code yielded invalid OAuth2 token")}
	}

	// Fetch information about the authenticated Google account.
	var googleInfo struct {
		UID string `json:"user_id"`
	}
	err = fetchAuth0UserInfo(r.Context(), token, &googleInfo)
	if err != nil {
		return err
	}
	googleRefreshToken, err := auth.FetchGoogleRefreshToken(r.Context(), googleInfo.UID)
	if err != nil {
		return fmt.Errorf("auth.FetchGoogleRefreshToken: %v", err)
	}

	// Link Google account to main account.
	actor := auth.ActorFromContext(r.Context())
	if err := auth.LinkAccount(r.Context(), actor.UID, "google-oauth2", googleInfo.UID); err != nil {
		return err
	}

	// Modify actor and write session cookie.
	actor.GoogleConnected = true
	actor.GoogleScopes = strings.Split(googleRefreshToken.Scope, ",")
	actor.GoogleRefreshToken = googleRefreshToken.Token
	if err := auth.StartNewSession(w, r, actor); err != nil {
		return err
	}

	// TODO: Add tracking info to return-to URL.
	returnToURL, err := url.Parse(returnTo)
	if err != nil {
		return err
	}

	http.Redirect(w, r, returnToURL.String(), http.StatusSeeOther)
	return nil
}
