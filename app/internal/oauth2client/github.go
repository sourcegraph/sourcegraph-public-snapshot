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

	"encoding/base64"

	"golang.org/x/oauth2"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/canonicalurl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/randstring"
)

var githubNonceCookiePath = router.Rel.URLTo(router.GitHubOAuth2Receive).Path

func auth0GitHubConfigWithRedirectURL(redirectURL string) *oauth2.Config {
	config := *auth.Auth0Config
	// RedirectURL is checked by Auth0 against a whitelist so it can't be spoofed.
	config.RedirectURL = redirectURL
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

	base := conf.AppURL
	// use X-App-Url header as base if available to make reverse proxies work
	if h := r.Header.Get("X-App-Url"); h != "" {
		if u, err := url.Parse(h); err == nil {
			base = u
		}
	}
	redirectURL := base.ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive))

	return GitHubOAuth2Initiate(w, r, scopes, redirectURL.String(), returnTo.String(), returnToNew.String())
}

type oauthCookie struct {
	Nonce       string
	RedirectURL string
	ReturnTo    string
	ReturnToNew string
}

func GitHubOAuth2Initiate(w http.ResponseWriter, r *http.Request, scopes []string, redirectURL, returnTo, returnToNew string) error {
	nonce := randstring.NewLen(32)
	cookie, err := json.Marshal(&oauthCookie{
		Nonce:       nonce,
		RedirectURL: redirectURL,
		ReturnTo:    returnTo,
		ReturnToNew: returnToNew,
	})
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "oauth",
		Value:   base64.URLEncoding.EncodeToString(cookie),
		Path:    "/",
		Expires: time.Now().Add(10 * time.Minute),
	})

	http.Redirect(w, r, auth0GitHubConfigWithRedirectURL(redirectURL).AuthCodeURL(nonce,
		oauth2.SetAuthURLParam("connection", "github"),
		oauth2.SetAuthURLParam("connection_scope", strings.Join(scopes, ",")),
	), http.StatusSeeOther)
	return nil
}

func ServeGitHubOAuth2Receive(w http.ResponseWriter, r *http.Request) (err error) {
	cookie := &oauthCookie{
		Nonce:       "", // the empty default value is not accepted unless impersonating
		RedirectURL: "",
		ReturnTo:    "/",
		ReturnToNew: "/",
	}
	if c, err := r.Cookie("oauth"); err == nil {
		cookieJSON, err := base64.URLEncoding.DecodeString(c.Value)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(cookieJSON, cookie); err != nil {
			return err
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth",
		Path:   "/",
		MaxAge: -1,
	})

	code := r.URL.Query().Get("code")
	token, err := auth0GitHubConfigWithRedirectURL(cookie.RedirectURL).Exchange(r.Context(), code)
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
		Name        string `json:"name"`
		Company     string `json:"company"`
		Location    string `json:"location"`
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
		expectedNonce := r.URL.Query().Get("state")
		if cookie.Nonce == "" || cookie.Nonce != expectedNonce {
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
				}
			}
		}
		if len(scopeOfToken) < len(mergedScope) {
			// The user has once granted us more permissions than we got with this token. Run oauth flow
			// again to fetch token with all permissions. This should be non-interactive.
			return GitHubOAuth2Initiate(w, r, mergedScope, cookie.RedirectURL, cookie.ReturnTo, cookie.ReturnToNew)
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

	// Write the session cookie.
	if err := auth.StartNewSession(w, r, actor); err != nil {
		return err
	}

	if firstTime {
		returnToNewURL, err := url.Parse(cookie.ReturnToNew)
		if err != nil {
			return err
		}
		q := returnToNewURL.Query()
		q.Set("tour", "signup")
		q.Set("_event", "SignupCompleted")
		q.Set("_signupChannel", "GitHubOAuth")
		q.Set("_githubAuthed", "true")
		q.Set("_githubCompany", info.Company)
		q.Set("_githubName", info.Name)
		q.Set("_githubLocation", info.Location)
		returnToNewURL.RawQuery = q.Encode()
		http.Redirect(w, r, returnToNewURL.String(), http.StatusSeeOther)
	} else {
		// Add tracking info to return-to URL.
		returnToURL, err := url.Parse(cookie.ReturnTo)
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
		q.Set("_githubCompany", info.Company)
		q.Set("_githubName", info.Name)
		q.Set("_githubLocation", info.Location)
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
