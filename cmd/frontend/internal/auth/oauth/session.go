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
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/session"
	"github.com/sourcegraph/sourcegraph/internal/actor"
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
			AnonymousUserID:            anonymousId,
			FirstSourceURL:             getCookie("first_page_seen_url"),
			LastSourceURL:              getCookie("last_page_seen_url"),
			LastPageSeenShort:          getCookie("last_page_seen_short"),
			LastPageSeenMid:            getCookie("last_page_seen_mid"),
			LastPageSeenLong:           getCookie("last_page_seen_long"),
			MostRecentReferrerUrl:      getCookie("most_recent_referrer_url"),
			MostRecentReferrerUrlShort: getCookie("most_recent_referrer_url_short"),
			MostRecentReferrerUrlMid:   getCookie("most_recent_referrer_url_mid"),
			MostRecentReferrerUrlLong:  getCookie("most_recent_referrer_url_long"),
			SignupSessionSourceURL:     getCookie("sourcegraphSignupSourceUrl"),
			SignupSessionReferrer:      getCookie("sourcegraphSignupReferrer"),
			SessionUTMCampaign:         getCookie("utm_campaign"),
			UtmCampaignShort:           getCookie("utm_campaign_short"),
			UtmCampaignMid:             getCookie("utm_campaign_mid"),
			UtmCampaignLong:            getCookie("utm_campaign_long"),
			SessionUTMSource:           getCookie("utm_source"),
			UtmSourceShort:             getCookie("utm_source_short"),
			UtmSourceMid:               getCookie("utm_source_mid"),
			UtmSourceLong:              getCookie("utm_source_long"),
			SessionUTMMedium:           getCookie("utm_medium"),
			UtmMediumShort:             getCookie("utm_medium_short"),
			UtmMediumMid:               getCookie("utm_medium_mid"),
			UtmMediumLong:              getCookie("utm_medium_long"),
			SessionUTMContent:          getCookie("utm_content"),
			UtmContentShort:            getCookie("utm_content_short"),
			UtmContentMid:              getCookie("utm_content_mid"),
			UtmContentLong:             getCookie("utm_content_long"),
			SessionUTMTerm:             getCookie("utm_term"),
			UtmTermShort:               getCookie("utm_term_short"),
			UtmTermMid:                 getCookie("utm_term_mid"),
			UtmTermLong:                getCookie("utm_term_long"),
			GoogleClickID:              getCookie("gclid"),
			MicrosoftClickID:           getCookie("msclkid"),
		})
		if err != nil {
			span.SetError(err)
			logger.Error("OAuth failed: error looking up or creating user from OAuth token.", log.Error(err), log.String("userErr", safeErrMsg))
			http.Error(w, safeErrMsg, http.StatusInternalServerError)

			if err = db.SecurityEventLogs().LogSecurityEvent(ctx, s.AuthFailedEventName(), r.URL.Path, uint32(actor.FromContext(ctx).UID), anonymousId, "BACKEND", nil); err != nil {
				logger.Warn("Error logging security event.", log.Error(err))
			}
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

		if err := db.SecurityEventLogs().LogSecurityEvent(ctx, s.AuthSucceededEventName(), r.URL.Path, uint32(user.ID), "", "BACKEND", nil); err != nil {
			logger.Warn("Error logging security event.", log.Error(err))
		}

		redirectURL := auth.AddPostAuthRedirectParametersToString(state.Redirect, newUserCreated, "OAuth::"+s.GetServiceID())
		http.Redirect(w, r, auth.SafeRedirectURL(redirectURL), http.StatusFound)
	})
}
