package oauth2client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/canonicalurl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
)

var githubNonceCookiePath = router.Rel.URLTo(router.GitHubOAuth2Receive).Path

func init() {
	internal.Handlers[router.GitHubOAuth2Initiate] = internal.Handler(serveGitHubOAuth2Initiate)
	internal.Handlers[router.GitHubOAuth2Receive] = internal.Handler(serveGitHubOAuth2Receive)
}

func auth0ConfigWithRedirectURL() *oauth2.Config {
	config := *auth.Auth0Config
	config.RedirectURL = conf.AppURL.ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive)).String()
	return &config
}

// serveGitHubOAuth2Initiate generates the OAuth2 authorize URL
// (including a nonce state value, also stored in a cookie) and
// redirects the client to that URL.
func serveGitHubOAuth2Initiate(w http.ResponseWriter, r *http.Request) error {
	returnTo, err := returnto.URLFromRequest(r)
	if err != nil {
		log15.Warn("Invalid return-to URL provided to OAuth2 flow initiation; ignoring.", "err", err)
	}

	// Remove UTM campaign params to avoid double
	// attribution. TODO(sqs): consider doing this on the frontend in
	// JS so we centralize usage analytics there.
	returnTo = canonicalurl.FromURL(returnTo)

	var scopes []string
	if s := r.URL.Query().Get("scopes"); s == "" {
		// if we have no scope, we upgrade the credential to the
		// minimum scope required, read access to email
		scopes = []string{"user:email"}
	} else {
		scopes = strings.Split(s, ",")
	}

	return oAuth2Initiate(w, r, scopes, returnTo.String())
}

func oAuth2Initiate(w http.ResponseWriter, r *http.Request, scopes []string, returnTo string) error {
	nonce := randstring.NewLen(32)
	http.SetCookie(w, &http.Cookie{
		Name:    "nonce",
		Value:   nonce,
		Path:    "",
		Expires: time.Now().Add(10 * time.Minute),
	})

	http.Redirect(w, r, auth0ConfigWithRedirectURL().AuthCodeURL(nonce+":"+returnTo,
		oauth2.SetAuthURLParam("connection", "github"),
		oauth2.SetAuthURLParam("connection_scope", strings.Join(scopes, ",")),
	), http.StatusSeeOther)
	return nil
}

func serveGitHubOAuth2Receive(w http.ResponseWriter, r *http.Request) (err error) {
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
	token, err := auth0ConfigWithRedirectURL().Exchange(r.Context(), code)
	if err != nil {
		return err
	}
	if !token.Valid() {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("exchanging auth code yielded invalid OAuth2 token")}
	}

	auth0Client := oauth2.NewClient(r.Context(), oauth2.StaticTokenSource(token))

	resp, err := auth0Client.Get("https://" + auth.Auth0Domain + "/userinfo")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var info struct {
		UID         string `json:"user_id"`
		Nickname    string `json:"nickname"`
		Picture     string `json:"picture"`
		Email       string `json:"email"`
		AppMetadata struct {
			GitHubScope []string `json:"github_scope"`
		} `json:"app_metadata"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return err
	}

	firstTime := len(info.AppMetadata.GitHubScope) == 0

	githubToken, err := auth.FetchGitHubToken(r.Context(), info.UID)
	if err != nil {
		return err
	}

	scopeOfToken := strings.Split(githubToken.Scope, ",")
	mergedScope := mergeScopes(scopeOfToken, info.AppMetadata.GitHubScope)
	if len(scopeOfToken) < len(mergedScope) {
		// The user has once granted us more permissions than we got with this token. Run oauth flow
		// again to fetch token with all permissions. This should be non-interactive.
		return oAuth2Initiate(w, r, mergedScope, returnTo)
	}
	if len(scopeOfToken) > len(info.AppMetadata.GitHubScope) {
		// Wohoo, we got more permissions. Remember in user database.
		if err := auth.SetAppMetadata(r.Context(), info.UID, "github_scope", scopeOfToken); err != nil {
			return err
		}
	}

	// Write cookie.
	if err := auth.StartNewSession(w, r, &auth.Actor{
		UID:             info.UID,
		Login:           info.Nickname,
		Email:           info.Email,
		AvatarURL:       info.Picture,
		GitHubConnected: true,
		GitHubScopes:    scopeOfToken,
		GitHubToken:     githubToken.Token,
	}); err != nil {
		return err
	}

	// Add tracking info to return-to URL.
	returnToURL, err := url.Parse(returnTo)
	if err != nil {
		return err
	}
	q := returnToURL.Query()
	if firstTime {
		q.Set("ob", "chrome")
		q.Set("_event", "SignupCompleted")
		q.Set("_signupChannel", "GitHubOAuth")
		q.Set("_githubAuthed", "true")
	} else {
		// Do not redirect a user while inside the onboarding flow.
		// This is accomplished by not removing the onboarding query params.
		if q.Get("ob") != "github" {
			q.Del("ob")
		}
		q.Set("_event", "CompletedGitHubOAuth2Flow")
		q.Set("_githubAuthed", "true")
	}
	returnToURL.RawQuery = q.Encode()

	http.Redirect(w, r, returnToURL.String(), http.StatusSeeOther)
	return nil
}

func mergeScopes(a, b []string) []string {
	m := make(map[string]struct{})
	for _, s := range a {
		m[s] = struct{}{}
	}
	for _, s := range b {
		m[s] = struct{}{}
	}

	var merged []string
	for s := range m {
		merged = append(merged, s)
	}
	sort.Strings(merged)
	return merged
}
