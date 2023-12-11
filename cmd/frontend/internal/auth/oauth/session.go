package oauth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	goauth2 "github.com/dghubble/gologin/v2/oauth2"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/cookie"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type SessionData struct {
	ID providers.ConfigID

	// Store only the oauth2.Token fields we need, to avoid hitting the ~4096-byte session data
	// limit.
	AccessToken string
	TokenType   string
}

type SessionIssuerHelper interface {
	GetOrCreateUser(ctx context.Context, token *oauth2.Token, hubSpotProps *hubspot.ContactProperties) (newUserCreated bool, actr *actor.Actor, safeErrMsg string, err error)
	DeleteStateCookie(w http.ResponseWriter, r *http.Request)
	SessionData(token *oauth2.Token) SessionData
	AuthSucceededEventName() database.SecurityEventName
	AuthFailedEventName() database.SecurityEventName
	GetServiceID() string
}

func SessionIssuer(logger log.Logger, db database.DB, s SessionIssuerHelper, sessionKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, ctx := trace.New(r.Context(), "oauth.SessionIssuer")
		defer span.End()

		// Scopes logger to family from trace.New
		logger := trace.Logger(ctx, logger)

		token, err := goauth2.TokenFromContext(ctx)
		if err != nil {
			span.SetError(err)
			logger.Error("OAuth failed: could not read token from context", log.Error(err))
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not read token from callback request.", http.StatusInternalServerError)
			return
		}

		expiryDuration := time.Duration(0)
		if token.Expiry != (time.Time{}) {
			expiryDuration = time.Until(token.Expiry)
		}
		if expiryDuration < 0 {
			span.SetError(err)
			logger.Error("OAuth failed: token was expired.")
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: OAuth token was expired.", http.StatusInternalServerError)
			return
		}

		encodedState, err := goauth2.StateFromContext(ctx)
		if err != nil {
			span.SetError(err)
			logger.Error("OAuth failed: could not get state from context.", log.Error(err))
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get OAuth state from context.", http.StatusInternalServerError)
			return
		}
		state, err := DecodeState(encodedState)
		if err != nil {
			span.SetError(err)
			logger.Error("OAuth failed: could not decode state.", log.Error(err))
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not get decode OAuth state.", http.StatusInternalServerError)
			return
		}
		logger = logger.With(
			log.String("ProviderID", state.ProviderID),
			log.String("Op", string(state.Op)),
		)
		span.SetAttributes(
			attribute.String("ProviderID", state.ProviderID),
			attribute.String("Op", string(state.Op)),
		)

		// Delete state cookie (no longer needed, will be stale if user logs out and logs back in within 120s)
		defer s.DeleteStateCookie(w, r)

		getCookie := func(name string) string {
			c, err := r.Cookie(name)
			if err != nil {
				return ""
			}
			return c.Value
		}
		anonymousId, _ := cookie.AnonymousUID(r)
		newUserCreated, actr, safeErrMsg, err := s.GetOrCreateUser(ctx, token, &hubspot.ContactProperties{
			AnonymousUserID:        anonymousId,
			FirstSourceURL:         getCookie("sourcegraphSourceUrl"),
			LastSourceURL:          getCookie("sourcegraphRecentSourceUrl"),
			OriginalReferrer:       getCookie("originalReferrer"),
			LastReferrer:           getCookie("Sg_referrer"),
			SignupSessionSourceURL: getCookie("sourcegraphSignupSourceUrl"),
			SignupSessionReferrer:  getCookie("sourcegraphSignupReferrer"),
			SessionUTMCampaign:     getCookie("Sg_utm_campaign"),
			SessionUTMSource:       getCookie("Sg_utm_source"),
			SessionUTMMedium:       getCookie("Sg_utm_medium"),
			SessionUTMContent:      getCookie("Sg_utm_content"),
			SessionUTMTerm:         getCookie("Sg_utm_term"),
			GoogleClickID:          getCookie("gclid"),
			MicrosoftClickID:       getCookie("msclkid"),
		})
		if err != nil {
			span.SetError(err)
			logger.Error("OAuth failed: error looking up or creating user from OAuth token.", log.Error(err), log.String("userErr", safeErrMsg))
			http.Error(w, safeErrMsg, http.StatusInternalServerError)
			db.SecurityEventLogs().LogEvent(ctx, &database.SecurityEvent{
				Name:            s.AuthFailedEventName(),
				URL:             r.URL.Path, // don't log query params w/ OAuth data
				AnonymousUserID: anonymousId,
				Source:          "BACKEND",
				Timestamp:       time.Now(),
			})
			return
		}

		user, err := db.Users().GetByID(ctx, actr.UID)
		if err != nil {
			span.SetError(err)
			logger.Error("OAuth failed: error retrieving user from database.", log.Error(err))
			http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not initiate session.", http.StatusInternalServerError)
			return
		}

		// Since we obtained a valid user from the OAuth token, we consider the login successful at this point
		ctx, err = session.SetActorFromUser(ctx, w, r, user, expiryDuration)
		if err != nil {
			span.SetError(err)
			logger.Error("OAuth failed: could not initiate session.", log.Error(err))
			http.Error(w, fmt.Sprintf("Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		db.SecurityEventLogs().LogEvent(ctx, &database.SecurityEvent{
			Name:      s.AuthSucceededEventName(),
			URL:       r.URL.Path, // don't log query params w/ OAuth data
			UserID:    uint32(user.ID),
			Source:    "BACKEND",
			Timestamp: time.Now(),
		})

		redirectURL := auth.AddPostAuthRedirectParametersToString(state.Redirect, newUserCreated, "OAuth::"+s.GetServiceID())
		http.Redirect(w, r, auth.SafeRedirectURL(redirectURL), http.StatusFound)
	})
}
