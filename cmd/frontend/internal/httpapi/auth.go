package httpapi

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/audit"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/cookie"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const authAuditEntity = "httpapi/auth"

// AccessTokenAuthMiddleware authenticates the user based on the
// token query parameter or the "Authorization" header.
func AccessTokenAuthMiddleware(db database.DB, baseLogger log.Logger, next http.Handler) http.Handler {
	baseLogger = baseLogger.Scoped("accessTokenAuth")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SCIM uses an auth token which is checked separately in the SCIM package.
		if strings.HasPrefix(r.URL.Path, "/.api/scim/v2") {
			next.ServeHTTP(w, r)
			return
		}

		// The license check handler uses a Bearer token and request body which
		// is checked in `productsubscription/license_check_handler.go`
		if strings.HasPrefix(r.URL.Path, "/.api/license/check") {
			next.ServeHTTP(w, r)
			return
		}

		logger := trace.Logger(r.Context(), baseLogger)

		w.Header().Add("Vary", "Authorization")

		var sudoUser string
		token := r.URL.Query().Get("token")

		if token == "" {
			// Handle token passed via basic auth (https://<token>@sourcegraph.com/foobar).
			basicAuthUsername, _, _ := r.BasicAuth()
			if basicAuthUsername != "" {
				token = basicAuthUsername
			}
		}

		if headerValue := r.Header.Get("Authorization"); headerValue != "" && token == "" {
			// Handle Authorization header
			var err error
			token, sudoUser, err = authz.ParseAuthorizationHeader(headerValue)
			if err != nil {
				if !envvar.SourcegraphDotComMode() && authz.IsUnrecognizedScheme(err) {
					// Ignore Authorization headers that we don't handle.
					// 🚨 SECURITY: sha256 the authorization header value so we redact it
					// while still retaining the ability to link it back to a token, assuming
					// the logs reader has the value in clear.
					var redactedValue string
					h := sha256.New()
					if _, err := io.WriteString(h, headerValue); err != nil {
						redactedValue = "[REDACTED]"
					} else {
						// for sake of identification, we only need around 10 characters
						redactedValue = fmt.Sprintf("sha256:%x", h.Sum(nil)[0:10])
					}
					// TODO: It is possible for the unrecognized header to be legitimate, in the case
					// of a customer setting up a HTTP header based authentication and decide to still
					// use the "Authorization" key.
					//
					// We should parse the configuration to see if that's the case and only log if it's
					// not defined over there.
					logger.Warn(
						"ignoring unrecognized Authorization header, passing it down to the next layer",
						log.String("redacted_value", redactedValue),
						log.Error(err),
					)
					next.ServeHTTP(w, r)
					return
				}

				// Report errors on malformed Authorization headers for schemes we do handle, to
				// make it clear to the client that their request is not proceeding with their
				// supplied credentials.
				audit.Log(r.Context(), logger, audit.Record{
					Entity: authAuditEntity,
					Action: "check_authorization_header",
					Fields: []log.Field{
						log.String("problem", "invalid Authorization header"),
						log.Error(err),
					},
				})
				http.Error(w, "Invalid Authorization header.", http.StatusUnauthorized)
				return
			}
		}

		if token != "" {
			if !(conf.AccessTokensAllow() == conf.AccessTokensAll || conf.AccessTokensAllow() == conf.AccessTokensAdmin) {
				// if conf.AccessTokensAllow() == conf.AccessTokensNone {
				http.Error(w, "Access token authorization is disabled.", http.StatusUnauthorized)
				return
			}

			// Validate access token.
			//
			// 🚨 SECURITY: It's important we check for the correct scopes to know what this token
			// is allowed to do.
			var requiredScope string
			if sudoUser == "" {
				requiredScope = authz.ScopeUserAll
			} else {
				requiredScope = authz.ScopeSiteAdminSudo
			}

			info, err := licensing.GetConfiguredProductLicenseInfo()
			if err != nil {
				http.Error(w, "Could not check license for access token authorization.", http.StatusInternalServerError)
				return
			}

			opts := database.TokenLookupOpts{
				RequiredScope: requiredScope,
				OnlyAdmin:     info.IsExpired(),
			}

			subjectUserID, err := db.AccessTokens().Lookup(r.Context(), token, opts)
			if err != nil {
				if err == database.ErrAccessTokenNotFound || errors.HasType(err, database.InvalidTokenError{}) {
					anonymousId, anonCookieSet := cookie.AnonymousUID(r)
					if !anonCookieSet {
						anonymousId = fmt.Sprintf("unknown user @ %s", time.Now()) // we don't have a reliable user identifier at the time of the failure
					}
					db.SecurityEventLogs().LogEvent(
						r.Context(),
						&database.SecurityEvent{
							Name:            database.SecurityEventAccessTokenInvalid,
							URL:             r.URL.RequestURI(),
							AnonymousUserID: anonymousId,
							Source:          "BACKEND",
							Timestamp:       time.Now(),
						},
					)

					http.Error(w, "Invalid access token.", http.StatusUnauthorized)
					return
				}

				logger.Error(
					"failed to look up access token",
					log.String("token", token),
					log.Error(err),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// FIXME: Can we find a way to do this only for SOAP users?
			soapCount, err := db.UserExternalAccounts().Count(
				r.Context(),
				database.ExternalAccountsListOptions{
					UserID:      subjectUserID,
					ServiceType: auth.SourcegraphOperatorProviderType,
				},
			)
			if err != nil {
				logger.Error(
					"failed to list user external accounts",
					log.Int32("subjectUserID", subjectUserID),
					log.Error(err),
				)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sourcegraphOperator := soapCount > 0

			// Determine the actor's user ID.
			var actorUserID int32
			if sudoUser == "" {
				actorUserID = subjectUserID
			} else {
				// 🚨 SECURITY: Confirm that the sudo token's subject is still a site admin, to
				// prevent users from retaining site admin privileges after being demoted.
				if err := auth.CheckUserIsSiteAdmin(r.Context(), db, subjectUserID); err != nil {
					logger.Error(
						"sudo access token's subject is not a site admin",
						log.Int32("subjectUserID", subjectUserID),
						log.Error(err),
					)

					args, err := json.Marshal(map[string]any{
						"subject_user_id": subjectUserID,
					})
					if err != nil {
						logger.Error(
							"failed to marshal JSON for security event log argument",
							log.String("eventName", string(database.SecurityEventAccessTokenSubjectNotSiteAdmin)),
							log.Error(err),
						)
						// OK to continue, we still want the security event log to be created
					}
					db.SecurityEventLogs().LogEvent(
						r.Context(),
						&database.SecurityEvent{
							Name:      database.SecurityEventAccessTokenSubjectNotSiteAdmin,
							URL:       r.URL.RequestURI(),
							UserID:    uint32(subjectUserID),
							Argument:  args,
							Source:    "BACKEND",
							Timestamp: time.Now(),
						},
					)

					http.Error(w, "The subject user of a sudo access token must be a site admin.", http.StatusForbidden)
					return
				}

				var tokenSubjectUserName string
				if tokenSubjectUser, err := db.Users().GetByID(r.Context(), subjectUserID); err == nil {
					tokenSubjectUserName = tokenSubjectUser.Username
				}

				// Sudo to the other user if this is a sudo token. We already checked that the token has
				// the necessary scope in the Lookup call above.
				user, err := db.Users().GetByUsername(r.Context(), sudoUser)
				if err != nil {
					audit.Log(r.Context(), logger, audit.Record{
						Entity: authAuditEntity,
						Action: "check_sudo_access",
						Fields: []log.Field{
							log.String("problem", "invalid username used with sudo access token"),
							log.String("sudoUser", sudoUser),
							log.Error(err),
						},
					})
					var message string
					if errcode.IsNotFound(err) {
						message = "Unable to sudo to nonexistent user."
					} else {
						message = "Unable to sudo to the specified user due to an unexpected error."
					}
					http.Error(w, message, http.StatusForbidden)
					return
				}
				actorUserID = user.ID
				logger.Debug(
					"HTTP request used sudo token",
					log.String("requestURI", r.URL.RequestURI()),
					log.Int32("tokenSubjectUserID", subjectUserID),
					log.Int32("actorUserID", actorUserID),
					log.String("actorUsername", user.Username),
				)

				args, err := json.Marshal(map[string]any{
					"sudo_user_id":            actorUserID,
					"sudo_user":               user.Username,
					"token_subject_user_id":   subjectUserID,
					"token_subject_user_name": tokenSubjectUserName,
				})
				if err != nil {
					logger.Error(
						"failed to marshal JSON for security event log argument",
						log.String("eventName", string(database.SecurityEventAccessTokenImpersonated)),
						log.String("sudoUser", sudoUser),
						log.Error(err),
					)
					// OK to continue, we still want the security event log to be created
				}
				db.SecurityEventLogs().LogEvent(
					actor.WithActor(
						r.Context(),
						&actor.Actor{
							UID:                 subjectUserID,
							SourcegraphOperator: sourcegraphOperator,
						},
					),
					&database.SecurityEvent{
						Name:      database.SecurityEventAccessTokenImpersonated,
						URL:       r.URL.RequestURI(),
						UserID:    uint32(subjectUserID),
						Argument:  args,
						Source:    "BACKEND",
						Timestamp: time.Now(),
					},
				)
			}

			r = r.WithContext(
				actor.WithActor(
					r.Context(),
					&actor.Actor{
						UID:                 actorUserID,
						SourcegraphOperator: sourcegraphOperator,
					},
				),
			)
		}

		next.ServeHTTP(w, r)
	})
}
