package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"encoding/base64"

	"golang.org/x/oauth2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
)

type oauthCookie struct {
	Nonce       string
	RedirectURL string
	ReturnTo    string
	ReturnToNew string
}

func auth0ConfigWithRedirectURL(redirectURL string) *oauth2.Config {
	config := *auth0.Config
	// RedirectURL is checked by Auth0 against a whitelist so it can't be spoofed.
	config.RedirectURL = redirectURL
	return &config
}

func ServeAuth0SignIn(w http.ResponseWriter, r *http.Request) (err error) {
	cookie := &oauthCookie{
		Nonce:       "",                   // the empty default value is not accepted unless impersonating
		RedirectURL: conf.AppURL.String(), // impersonation does not allow this to be empty
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
	token, err := auth0ConfigWithRedirectURL(cookie.RedirectURL).Exchange(r.Context(), code)
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
			DidLoginBefore bool `json:"did_login_before"`
		} `json:"app_metadata"`
		Identities []struct {
			Connection string          `json:"connection"`
			UserID     json.RawMessage `json:"user_id"`
		} `json:"identities"`
		Impersonated bool `json:"impersonated"`
	}
	err = fetchAuth0UserInfo(r.Context(), token, &info)
	if err != nil {
		return err
	}

	actor := &actor.Actor{
		UID:             info.UID,
		Login:           info.Nickname,
		Email:           info.Email,
		AvatarURL:       info.Picture,
		GitHubConnected: false, // TODO: Remove
	}

	// Write the session cookie.
	if err := session.StartNewSession(w, r, actor); err != nil {
		return err
	}

	eventLabel := "CompletedAuth0SignIn"
	if !info.AppMetadata.DidLoginBefore {
		eventLabel = "SignupCompleted"
	}

	// Track user data in GCS
	if r.UserAgent() != "Sourcegraph e2etest-bot" {
		go tracking.TrackUser(actor, eventLabel)
	}

	returnTo := r.URL.Query().Get("return-to")
	if returnTo == "" {
		returnTo = "/"
	}

	if !info.AppMetadata.DidLoginBefore {
		if err := auth0.SetAppMetadata(r.Context(), info.UID, "did_login_before", true); err != nil {
			return err
		}
		returnToNewURL, err := url.Parse(cookie.ReturnToNew)
		if err != nil {
			return err
		}
		q := returnToNewURL.Query()
		q.Set("_event", eventLabel)
		returnToNewURL.RawQuery = q.Encode()
		http.Redirect(w, r, returnTo, http.StatusSeeOther)
	} else {
		// Add tracking info to return-to URL.
		returnToURL, err := url.Parse(cookie.ReturnTo)
		if err != nil {
			return err
		}
		q := returnToURL.Query()
		q.Set("_event", eventLabel)
		returnToURL.RawQuery = q.Encode()
		http.Redirect(w, r, returnTo, http.StatusSeeOther)
	}

	return nil
}

// fetchAuth0UserInfo fetches Auth0 user info for token into v.
func fetchAuth0UserInfo(ctx context.Context, token *oauth2.Token, v interface{}) error {
	auth0Client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := auth0Client.Get("https://" + auth0.Domain + "/userinfo")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(&v)
}
