package oauth

import (
	"context"
	"net/http"
	"time"

	goauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"golang.org/x/oauth2"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type sessionData struct {
	ID auth.ProviderConfigID

	// Store only the oauth2.Token fields we need, to avoid hitting the ~4096-byte session data
	// limit.
	AccessToken string
	TokenType   string
}

func SessionIssuer(
	sessionKey, serviceType, serviceID, clientID string,
	getOrCreateUser func(ctx context.Context, serviceType, serviceID, clientID string, token *oauth2.Token) (actr *actor.Actor, safeErrMsg string, err error),
	deleteStateCookie func(w http.ResponseWriter),
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, err := goauth2.TokenFromContext(ctx)
		if err != nil {
			log15.Error("OAuth failed: could not read token from context", "error", err)
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not read token from callback request.", http.StatusInternalServerError)
			return
		}

		actr, safeErrMsg, err := getOrCreateUser(ctx, serviceType, serviceID, clientID, token)
		if err != nil {
			log15.Error("OAuth failed: error looking up or creating user from OAuth token.", "error", err, "userErr", safeErrMsg)
			http.Error(w, safeErrMsg, http.StatusInternalServerError)
			return
		}

		expiryDuration := time.Duration(0)
		if token.Expiry != (time.Time{}) {
			expiryDuration = time.Until(token.Expiry)
		}
		if expiryDuration < 0 {
			log15.Error("OAuth failed: token was expired.")
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: OAuth token was expired.", http.StatusInternalServerError)
			return
		}
		if err := session.SetActor(w, r, actr, expiryDuration); err != nil { // TODO: test session expiration
			log15.Error("OAuth failed: could not initiate session.", "error", err)
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not initiate session.", http.StatusInternalServerError)
			return
		}

		encodedState, err := goauth2.StateFromContext(ctx)
		if err != nil {
			log15.Error("OAuth failed: could not get state from context.", "error", err)
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get OAuth state from context.", http.StatusInternalServerError)
			return
		}
		state, err := DecodeState(encodedState)
		if err != nil {
			log15.Error("OAuth failed: could not decode state.", "error", err)
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get decode OAuth state.", http.StatusInternalServerError)
			return
		}

		// Set session data, which can be used by logout handler
		data := sessionData{
			ID: auth.ProviderConfigID{
				ID:   serviceID,
				Type: serviceType,
			},
			AccessToken: token.AccessToken,
			TokenType:   token.Type(),
			// TODO(beyang): store and use refresh token to auto-refresh sessions
		}
		if err := session.SetData(w, r, sessionKey, data); err != nil {
			// It's not fatal if this fails. It just means we won't be able to sign the user out of
			// the OP.
			log15.Warn("Failed to set OAuth session data. The session is still secure, but Sourcegraph will be unable to revoke the user's token or redirect the user to the end-session endpoint after the user signs out of Sourcegraph.", "error", err)
		}

		// Delete state cookie (no longer needed, while be stale if user logs out and logs back in within 120s)
		deleteStateCookie(w)

		http.Redirect(w, r, auth.SafeRedirectURL(state.Redirect), http.StatusFound)
	})
}
