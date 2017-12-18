package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"encoding/base64"

	"github.com/mattbaird/gochimp"
	"golang.org/x/oauth2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/invite"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tracking"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

type oauthCookie struct {
	Nonce       string
	RedirectURL string
	ReturnTo    string
	ReturnToNew string
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

func nativeAuthID(email string) string {
	return fmt.Sprintf("native:%s", email)
}

func serveSignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	displayName := creds.DisplayName
	if displayName == "" {
		displayName = creds.Username
	}

	// Create user
	emailCodeBytes := make([]byte, 20)
	if _, err := rand.Read(emailCodeBytes); err != nil {
		http.Error(w, "could not generate email code", http.StatusInternalServerError)
		return
	}
	emailCode := base64.StdEncoding.EncodeToString(emailCodeBytes)
	usr, err := store.Users.Create(r.Context(), nativeAuthID(creds.Email), creds.Email, creds.Username, displayName, "native", nil, creds.Password, emailCode)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not create user %s", creds.Username), http.StatusInternalServerError)
		return
	}
	actor := &actor.Actor{
		UID:   usr.Auth0ID,
		Login: usr.Username,
		Email: usr.Email,
	}

	// Send verify email
	q := make(url.Values)
	q.Set("code", emailCode)
	verifyLink := globals.AppURL.String() + router.Rel.URLTo(router.VerifyEmail).Path + "?" + q.Encode()
	notif.SendMandrillTemplate(&notif.EmailConfig{
		Template:  "verify-email",
		FromEmail: "noreply@sourcegraph.com",
		ToEmail:   creds.Email,
		Subject:   "Verify your email on Sourcegraph Server",
	}, []gochimp.Var{}, []gochimp.Var{{Name: "VERIFY_URL", Content: verifyLink}})

	// Write the session cookie
	if session.StartNewSession(w, r, actor, 0); err != nil {
		http.Error(w, "could not create new user session", http.StatusInternalServerError)
		return
	}
}

// serveSignIn2 serves a native-auth endpoint
func serveSignIn2(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StatusBadRequest)
		return
	}
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "could not decode request body", http.StatusBadRequest)
		return
	}

	// Validate user
	authID := nativeAuthID(creds.Email)
	usr, err := store.Users.GetByAuth0ID(r.Context(), authID)
	if err != nil {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}
	actor := &actor.Actor{
		UID:   usr.Auth0ID,
		Login: usr.Username,
		Email: usr.Email,
	}

	// Write the session cookie
	if session.StartNewSession(w, r, actor, 0); err != nil {
		http.Error(w, "could not create new user session", http.StatusInternalServerError)
		return
	}

	// Track user data in GCS
	eventLabel := "CompletedNativeSignIn"
	if r.UserAgent() != "Sourcegraph e2etest-bot" {
		go tracking.TrackUser(actor, eventLabel)
	}
}

func serveVerifyEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	verifyCode := r.URL.Query().Get("code")
	actr := actor.FromContext(ctx)
	if !actr.IsAuthenticated() {
		redirectTo := r.URL.String()
		q := make(url.Values)
		q.Set("returnTo", redirectTo)
		http.Redirect(w, r, "/sign-in?"+q.Encode(), http.StatusFound)
		return
	}
	usr, err := store.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		http.Error(w, "Could not get current user.", http.StatusUnauthorized)
		return
	}
	if usr.Verified {
		http.Error(w, fmt.Sprintf("User %q already verified", usr.Email), http.StatusBadRequest)
		return
	}
	verified, err := store.Users.ValidateEmail(ctx, usr.ID, verifyCode)
	if err != nil {
		http.Error(w, "Unexpected error when verifying user.", http.StatusInternalServerError)
		return
	}
	if !verified {
		http.Error(w, "Could not verify user. Code did not match email.", http.StatusUnauthorized)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusFound)
}

// DEPRECATED
func auth0ConfigWithRedirectURL(redirectURL string) *oauth2.Config {
	config := *auth0.Config
	// RedirectURL is checked by Auth0 against a whitelist so it can't be spoofed.
	config.RedirectURL = redirectURL
	return &config
}

// DEPRECATED
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
		dbUser, userCreateErr = store.Users.Create(r.Context(), info.UserID, info.Email, username, displayName, "", &info.Picture, "", "")
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
