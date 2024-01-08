package sourcegraphoperator

import (
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	internalauth "github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// All Sourcegraph Operator endpoints are under this path prefix.
const authPrefix = auth.AuthURLPrefix + "/" + internalauth.SourcegraphOperatorProviderType

// Middleware is middleware for Sourcegraph Operator authentication, adding
// endpoints under the auth path prefix ("/.auth") to enable the login flow and
// requiring login for all other endpoints.
//
// 🚨SECURITY: See docstring of the openidconnect.Middleware for security details
// because the Sourcegraph Operator authentication provider is a wrapper of the
// OpenID Connect authentication provider.
func Middleware(db database.DB) *auth.Middleware {
	return &auth.Middleware{
		API: func(next http.Handler) http.Handler {
			// Pass through to the next handler for API requests.
			return next
		},
		App: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Delegate to the Sourcegraph Operator authentication handler.
				if strings.HasPrefix(r.URL.Path, authPrefix+"/") {
					authHandler(db)(w, r)
					return
				}

				next.ServeHTTP(w, r)
			})
		},
	}
}

// SessionKey is the key of the key-value pair in a user session for the
// Sourcegraph Operator authentication provider.
const SessionKey = "soap@0"

const usernamePrefix = "sourcegraph-operator-"

func authHandler(db database.DB) func(w http.ResponseWriter, r *http.Request) {
	logger := log.Scoped(internalauth.SourcegraphOperatorProviderType + ".authHandler")
	return func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimPrefix(r.URL.Path, authPrefix) {
		case "/login": // Endpoint that starts the Authentication Request Code Flow.
			p, safeErrMsg, err := openidconnect.GetProviderAndRefresh(r.Context(), r.URL.Query().Get("pc"), GetOIDCProvider)
			if err != nil {
				logger.Error("failed to get provider", log.Error(err))
				http.Error(w, safeErrMsg, http.StatusInternalServerError)
				return
			}
			openidconnect.RedirectToAuthRequest(w, r, p, r.URL.Query().Get("redirect"))
			return

		case "/callback": // Endpoint for the OIDC Authorization Response, see http://openid.net/specs/openid-connect-core-1_0.html#AuthResponse.
			result, safeErrMsg, errStatus, err := openidconnect.AuthCallback(db, r, usernamePrefix, GetOIDCProvider)
			if err != nil {
				logger.Error("failed to authenticate with Sourcegraph Operator", log.Error(err))
				http.Error(w, safeErrMsg, errStatus)
				return
			}

			p, ok := providers.GetProviderByConfigID(
				providers.ConfigID{
					Type: internalauth.SourcegraphOperatorProviderType,
					ID:   internalauth.SourcegraphOperatorProviderType,
				},
			).(*provider)
			if !ok {
				logger.Error(
					"failed to get Sourcegraph Operator authentication provider",
					log.Error(errors.Errorf("no authentication provider found with ID %q", internalauth.SourcegraphOperatorProviderType)),
				)
				http.Error(w, "Misconfigured authentication provider.", http.StatusInternalServerError)
				return
			}

			extAccts, err := db.UserExternalAccounts().List(
				r.Context(),
				database.ExternalAccountsListOptions{
					UserID: result.User.ID,
					LimitOffset: &database.LimitOffset{
						Limit: 2,
					},
				},
			)
			if err != nil {
				logger.Error("failed list user external accounts", log.Error(err))
				http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not list user external accounts.", http.StatusInternalServerError)
				return
			}

			var expiry time.Duration
			// If the "sourcegraph-operator" (SOAP) is the only external account associated
			// with the user, that means the user is a pure Sourcegraph Operator which should
			// have designated and aggressive session expiry - unless that account is designated
			// as a service account. However, because service accounts are not "real" users and
			// cannot log in through the user interface (instead, we provision access entirely
			// via API tokens), we do not add special handling here to avoid deleting service
			// accounts.
			if len(extAccts) == 1 && extAccts[0].ServiceType == internalauth.SourcegraphOperatorProviderType {
				// The user session will only live at most for the remaining duration from the
				// "users.created_at" compared to the current time.
				//
				// For example, if a Sourcegraph operator user account is created at
				// "2022-10-10T10:10:10Z" and the configured lifecycle duration is one hour, this
				// account will be deleted as early as "2022-10-10T11:10:10Z", which means:
				//   - Upon creation of an account, the session lives for an hour.
				//   - If the same operator signs out and signs back in again after 10 minutes,
				//       the second session only lives for 50 minutes.
				expiry = time.Until(result.User.CreatedAt.Add(sourcegraphoperator.LifecycleDuration(p.config.LifecycleDuration)))
				if expiry <= 0 {
					// Let's do a proactive hard delete since the background worker hasn't caught up

					// Help exclude Sourcegraph operator related events from analytics
					ctx := actor.WithActor(
						r.Context(),
						&actor.Actor{
							SourcegraphOperator: true,
						},
					)
					err = db.Users().HardDelete(ctx, result.User.ID)
					if err != nil {
						logger.Error("failed to proactively clean up the expire user account", log.Error(err))
					}

					http.Error(w, "The retrieved user account lifecycle has already expired, please re-authenticate.", http.StatusUnauthorized)
					return
				}
			}

			act := &actor.Actor{
				UID:                 result.User.ID,
				SourcegraphOperator: true,
			}
			err = session.SetActor(w, r, act, expiry, result.User.CreatedAt)
			if err != nil {
				logger.Error("failed to authenticate with Sourcegraph Operator", log.Error(errors.Wrap(err, "initiate session")))
				http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not initiate session.", http.StatusInternalServerError)
				return
			}

			// NOTE: It is important to wrap the request context with the correct actor and
			// use it onwards to be able to mark all generated event logs with
			// `"sourcegraph_operator": true`.
			ctx := actor.WithActor(r.Context(), act)

			if err = session.SetData(w, r, SessionKey, result.SessionData); err != nil {
				// It's not fatal if this fails. It just means we won't be able to sign the user
				// out of the OP.
				logger.Warn(
					"failed to set Sourcegraph Operator session data",
					log.String("message", "The session is still secure, but Sourcegraph will be unable to revoke the user's token or redirect the user to the end-session endpoint after the user signs out of Sourcegraph."),
					log.Error(err),
				)
			} else {
				arg := map[string]any{
					"session_expiry_seconds": int64(expiry.Seconds()),
				}
				if err := db.SecurityEventLogs().LogSecurityEvent(ctx, database.SecurityEventNameSignInSucceeded, r.URL.Path, uint32(act.UID), "", "BACKEND", arg); err != nil {
					logger.Warn("Error logging security event", log.Error(err))
				}
			}

			if !result.User.SiteAdmin {
				err = db.Users().SetIsSiteAdmin(ctx, result.User.ID, true)
				if err != nil {
					logger.Error("failed to update Sourcegraph Operator as site admin", log.Error(err))
					http.Error(w, "Authentication failed. Try signing in again (and clearing cookies for the current site). The error was: could not set as site admin.", http.StatusInternalServerError)
					return
				}
			}

			// 🚨 SECURITY: Call auth.SafeRedirectURL to avoid the open-redirect vulnerability.
			http.Redirect(w, r, auth.SafeRedirectURL(result.Redirect), http.StatusFound)

		default:
			http.NotFound(w, r)
		}
	}
}
