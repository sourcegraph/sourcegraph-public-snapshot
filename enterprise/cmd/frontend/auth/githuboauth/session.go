package githuboauth

import (
	"net/http"
	"time"

	"github.com/dghubble/gologin/github"
	goauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

const sessionKey = "githuboauth@0"

type sessionData struct {
	ID auth.ProviderConfigID

	// Store only the oauth2.Token fields we need, to avoid hitting the ~4096-byte session data
	// limit.
	AccessToken string
	TokenType   string
}

func issueSession(p *provider, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token, err := goauth2.TokenFromContext(ctx)
	if err != nil {
		log15.Error("GitHub OAuth auth failed: could not read token from context", "error", err)
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not read token from callback request.", http.StatusInternalServerError)
		return
	}
	githubUser, err := github.UserFromContext(ctx)
	if err != nil {
		log15.Error("GitHub OAuth auth failed: could not read user from context", "error", err)
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not read GitHub user from callback request.", http.StatusInternalServerError)
		return
	}

	actr, safeErrMsg, err := getOrCreateUser(ctx, p, githubUser, token)
	if err != nil {
		log15.Error("GitHub OAuth failed: error looking up or creating GitHub user.", "error", err, "userErr", safeErrMsg)
		http.Error(w, safeErrMsg, http.StatusInternalServerError)
		return
	}

	var expiryDuration time.Duration = 0
	if token.Expiry != (time.Time{}) {
		expiryDuration = time.Until(token.Expiry)
	}
	if expiryDuration < 0 {
		log15.Error("GitHub OAuth failed: token was expired.")
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: OAuth token was expired.", http.StatusInternalServerError)
		return
	}
	if err := session.SetActor(w, r, actr, expiryDuration); err != nil { // TODO: test session expiration
		log15.Error("GitHub OAuth failed: could not initiate session.", "error", err)
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not initiate session.", http.StatusInternalServerError)
		return
	}

	encodedState, err := goauth2.StateFromContext(ctx)
	if err != nil {
		log15.Error("GitHub OAuth failed: could not get state from context.", "error", err)
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get OAuth state from context.", http.StatusInternalServerError)
		return
	}
	state, err := DecodeState(encodedState)
	if err != nil {
		log15.Error("GitHub OAuth failed: could not decode state.", "error", err)
		http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get decode OAuth state.", http.StatusInternalServerError)
		return
	}

	// Delete state cookie (no longer needed, while be stale if user logs out and logs back in within 120s)
	stateConfig := getStateConfig()
	stateConfig.MaxAge = -1
	http.SetCookie(w, newCookie(stateConfig, ""))

	data := sessionData{
		ID:          p.ConfigID(),
		AccessToken: token.AccessToken,
		TokenType:   token.Type(),
		// TODO(beyang): store and use refresh token to auto-refresh sessions
	}
	if err := session.SetData(w, r, sessionKey, data); err != nil {
		// It's not fatal if this fails. It just means we won't be able to sign the user out of
		// the OP.
		log15.Warn("Failed to set GitHub OAuth session data. The session is still secure, but Sourcegraph will be unable to revoke the user's token or redirect the user to the end-session endpoint after the user signs out of Sourcegraph.", "error", err)
	}
	http.Redirect(w, r, auth.SafeRedirectURL(state.Redirect), http.StatusFound)
}
