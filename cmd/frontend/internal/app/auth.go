package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"encoding/base64"

	"golang.org/x/oauth2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/invite"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
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
		Nonce:       "",                      // the empty default value is not accepted unless impersonating
		RedirectURL: globals.AppURL.String(), // impersonation does not allow this to be empty
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

	info := auth0.User{}
	err = fetchAuth0UserInfo(r.Context(), token, &info)
	if err != nil {
		return err
	}

	username := r.URL.Query().Get("username")
	displayName := r.URL.Query().Get("displayName")
	if displayName == "" {
		displayName = username
	}

	dbUser, err := store.Users.GetByEmail(r.Context(), info.Email)
	if err != nil {
		if _, ok := err.(store.ErrUserNotFound); !ok {
			// Return all but "user not found" errors;
			// handle those by creating a db row.
			return err
		}
	}
	var userCreateErr error
	if dbUser == nil {
		// Create the user in our DB if the user just signed up via Auth0. There is a TOCTTOU
		// bug here; their username may no longer be available. Because this is a rare case and
		// we are removing Auth0 soon, we ignore it.
		dbUser, userCreateErr = store.Users.Create(r.Context(), info.UserID, info.Email, username, displayName, "", &info.Picture)
		if userCreateErr != nil {
			return err
		}
	}

	actor := &actor.Actor{
		UID:             info.UserID,
		Login:           username,
		Email:           info.Email,
		AvatarURL:       info.Picture,
		GitHubConnected: false, // TODO: Remove
	}

	// Write the session cookie
	if err := session.StartNewSession(w, r, actor, 0); err != nil {
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

	returnTo := r.URL.Query().Get("returnTo")
	if returnTo == "" {
		returnTo = "/"
	}

	userToken := r.URL.Query().Get("token")
	if dbUser != nil && userToken != "" {
		// Add editor beta tag for a new user that signs up, if they have been invited to an org.
		_, err := addEditorBetaTag(r.Context(), dbUser, userToken)
		if err != nil {
			return err
		}
	}

	if !info.AppMetadata.DidLoginBefore {
		if err := auth0.SetAppMetadata(r.Context(), info.UserID, "did_login_before", true); err != nil {
			return err
		}
		returnToURL, err := url.Parse(returnTo)
		if err != nil {
			return err
		}
		q := returnToURL.Query()
		q.Set("_event", eventLabel)
		returnToURL.RawQuery = q.Encode()
		http.Redirect(w, r, returnToURL.String(), http.StatusSeeOther)
	} else {
		// Add tracking info to returnTo URL.
		returnToURL, err := url.Parse(returnTo)
		if err != nil {
			return err
		}
		q := returnToURL.Query()
		q.Set("_event", eventLabel)
		returnToURL.RawQuery = q.Encode()
		http.Redirect(w, r, returnToURL.String(), http.StatusSeeOther)
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

func addEditorBetaTag(ctx context.Context, user *sourcegraph.User, tokenString string) (*sourcegraph.UserTag, error) {
	// 🚨 SECURITY: verify that the token is valid before adding editor-beta tag
	_, err := invite.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}
	return store.UserTags.CreateIfNotExists(ctx, user.ID, "editor-beta")
}
