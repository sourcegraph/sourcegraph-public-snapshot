package oauth2client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/canonicalurl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

var githubNonceCookiePath = router.Rel.URLTo(router.GitHubOAuth2Receive).Path

func auth0GitHubConfigWithRedirectURL() *oauth2.Config {
	config := *auth.Auth0Config
	config.RedirectURL = conf.AppURL.ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive)).String()
	return &config
}

// ServeGitHubOAuth2Initiate generates the OAuth2 authorize URL
// (including a nonce state value, also stored in a cookie) and
// redirects the client to that URL.
func ServeGitHubOAuth2Initiate(w http.ResponseWriter, r *http.Request) error {
	returnTo, err := returnto.URLFromRequest(r, "return-to")
	if err != nil {
		log15.Warn("Invalid return-to URL provided to OAuth2 flow initiation; ignoring.", "err", err)
	}

	// Remove UTM campaign params to avoid double
	// attribution. TODO(sqs): consider doing this on the frontend in
	// JS so we centralize usage analytics there.
	returnTo = canonicalurl.FromURL(returnTo)

	returnToNew, err := returnto.URLFromRequest(r, "new-user-return-to")
	if err != nil {
		log15.Warn("Invalid new-user-return-to URL provided to OAuth2 flow initiation; ignoring.", "err", err)
	}

	var scopes []string
	if s := r.URL.Query().Get("scopes"); s == "" {
		// if we have no scope, we upgrade the credential to the
		// minimum scope required, read access to email
		scopes = []string{"user:email"}
	} else {
		scopes = strings.Split(s, ",")
	}

	return githubOAuth2Initiate(w, r, scopes, returnTo.String(), returnToNew.String())
}

func githubOAuth2Initiate(w http.ResponseWriter, r *http.Request, scopes []string, returnTo string, returnToNew string) error {
	nonce := randstring.NewLen(32)
	http.SetCookie(w, &http.Cookie{
		Name:    "nonce",
		Value:   nonce,
		Path:    "",
		Expires: time.Now().Add(10 * time.Minute),
	})

	http.Redirect(w, r, auth0GitHubConfigWithRedirectURL().AuthCodeURL(nonce+":"+returnTo+":"+returnToNew,
		oauth2.SetAuthURLParam("connection", "github"),
		oauth2.SetAuthURLParam("connection_scope", strings.Join(scopes, ",")),
	), http.StatusSeeOther)
	return nil
}

func ServeGitHubOAuth2Receive(w http.ResponseWriter, r *http.Request) (err error) {
	expectedNonceCookie := ""
	returnTo := "/"
	returnToNew := "/"
	if parts := strings.SplitN(r.URL.Query().Get("state"), ":", 3); len(parts) == 3 {
		expectedNonceCookie = parts[0]
		returnTo = parts[1]
		returnToNew = parts[2]
	}

	code := r.URL.Query().Get("code")
	token, err := auth0GitHubConfigWithRedirectURL().Exchange(r.Context(), code)
	if err != nil {
		return err
	}
	if !token.Valid() {
		return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("exchanging auth code yielded invalid OAuth2 token")}
	}

	var info struct {
		UID         string `json:"user_id"`
		Nickname    string `json:"nickname"`
		Picture     string `json:"picture"`
		Email       string `json:"email"`
		AppMetadata struct {
			GitHubScope               []string `json:"github_scope"`
			GitHubAccessTokenOverride string   `json:"github_access_token_override"`
		} `json:"app_metadata"`
		Identities []struct {
			Connection string          `json:"connection"`
			UserID     json.RawMessage `json:"user_id"` // Defer decoding because the type is int for GitHub, but string for Google.
		} `json:"identities"`
		Impersonated bool `json:"impersonated"`
	}
	err = fetchAuth0UserInfo(r.Context(), token, &info)
	if err != nil {
		return err
	}

	if !info.Impersonated { // impersonation has no state parameter, so don't check nonce
		nonceCookie, err := r.Cookie("nonce")
		if err != nil {
			return err
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "nonce",
			Path:   "/",
			MaxAge: -1,
		})
		if nonceCookie.Value == "" || expectedNonceCookie != nonceCookie.Value {
			return &errcode.HTTPErr{Status: http.StatusForbidden, Err: errors.New("invalid state")}
		}
	}

	firstTime := len(info.AppMetadata.GitHubScope) == 0

	actor := &auth.Actor{
		UID:             info.UID,
		Login:           info.Nickname,
		Email:           info.Email,
		AvatarURL:       info.Picture,
		GitHubConnected: true,
	}

	if info.AppMetadata.GitHubAccessTokenOverride == "" {
		githubToken, err := auth.FetchGitHubToken(r.Context(), info.UID)
		if err != nil {
			return fmt.Errorf("auth.FetchGitHubToken: %v", err)
		}

		scopeOfToken := strings.Split(githubToken.Scope, ",")
		mergedScope := mergeScopes(scopeOfToken, info.AppMetadata.GitHubScope)
		if firstTime {
			// try copying legacy scope
			for _, identity := range info.Identities {
				if identity.Connection == "github" {
					var githubUserID int
					err := json.Unmarshal(identity.UserID, &githubUserID)
					if err != nil {
						log15.Warn(`Connection is "github", but UserID type isn't int; ignoring.`, "UserID", identity.UserID, "err", err)
						continue
					}
					if legacyScope := backend.LegacyGitHubScope(githubUserID); len(legacyScope) > 0 {
						firstTime = false
						mergedScope = mergeScopes(mergedScope, legacyScope)
					}
				}
			}
		}
		if len(scopeOfToken) < len(mergedScope) {
			// The user has once granted us more permissions than we got with this token. Run oauth flow
			// again to fetch token with all permissions. This should be non-interactive.
			return githubOAuth2Initiate(w, r, mergedScope, returnTo, returnToNew)
		}
		if len(scopeOfToken) > len(info.AppMetadata.GitHubScope) {
			// Wohoo, we got more permissions. Remember in user database.
			if err := auth.SetAppMetadata(r.Context(), info.UID, "github_scope", scopeOfToken); err != nil {
				return err
			}
		}

		actor.GitHubScopes = scopeOfToken
		actor.GitHubToken = githubToken.Token
	} else {
		actor.GitHubScopes = []string{"read:org", "repo", "user:email"}
		actor.GitHubToken = info.AppMetadata.GitHubAccessTokenOverride
	}

	var googleConnected bool
	for _, identity := range info.Identities {
		if identity.Connection == "google-oauth2" {
			googleConnected = true
			break
		}
	}
	if googleConnected {
		googleRefreshToken, err := auth.FetchGoogleRefreshToken(r.Context(), info.UID)
		if err != nil {
			return fmt.Errorf("auth.FetchGoogleRefreshToken: %v", err)
		}

		actor.GoogleConnected = true
		actor.GoogleScopes = strings.Split(googleRefreshToken.Scope, ",")
	}
	// Write the session cookie.
	if err := auth.StartNewSession(w, r, actor); err != nil {
		return err
	}

	if firstTime {
		returnToNewURL, err := url.Parse(returnToNew)
		if err != nil {
			return err
		}
		q := returnToNewURL.Query()
		q.Set("tour", "signup")
		q.Set("_event", "SignupCompleted")
		q.Set("_signupChannel", "GitHubOAuth")
		q.Set("_githubAuthed", "true")
		returnToNewURL.RawQuery = q.Encode()
		http.Redirect(w, r, returnToNewURL.String(), http.StatusSeeOther)
	} else {
		// Add tracking info to return-to URL.
		returnToURL, err := url.Parse(returnTo)
		if err != nil {
			return err
		}
		q := returnToURL.Query()
		// Do not redirect a user while inside the onboarding flow.
		// This is accomplished by not removing the onboarding query params.
		if q.Get("ob") != "github" {
			q.Del("ob")
		}
		q.Set("_event", "CompletedGitHubOAuth2Flow")
		q.Set("_githubAuthed", "true")
		returnToURL.RawQuery = q.Encode()
		http.Redirect(w, r, returnToURL.String(), http.StatusSeeOther)
	}

	return nil
}

// fetchAuth0UserInfo fetches Auth0 user info for token into v.
func fetchAuth0UserInfo(ctx context.Context, token *oauth2.Token, v interface{}) error {
	auth0Client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := auth0Client.Get("https://" + auth.Auth0Domain + "/userinfo")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(&v)
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
