// Package session implements a redis backed user sessions HTTP middleware.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/boj/redistore"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const SignOutCookie = "sg-signout"

var (
	sessionCookieKey = env.Get("SRC_SESSION_COOKIE_KEY", "", "secret key used for securing the session cookies")
)

// defaultExpiryPeriod is the default session expiry period (if none is specified explicitly): 90 days.
const defaultExpiryPeriod = 90 * 24 * time.Hour

// cookieName is the name of the HTTP cookie that stores the session ID.
const cookieName = "sgs"

// sessionInfo is the information we store in the session. The gorilla/sessions library doesn't appear to
// enforce the maxAge field in its session store implementations, so we include the expiry here.
type sessionInfo struct {
	Actor         *actor.Actor  `json:"actor"`
	LastActive    time.Time     `json:"lastActive"`
	ExpiryPeriod  time.Duration `json:"expiryPeriod"`
	UserCreatedAt time.Time     `json:"userCreatedAt"`
}

// sessionsStore wraps another sessions.Store to dynamically set the values
// of the session.Options.Secure and session.Options.SameSite fields to what
// is returned by the secure closure at invocation time.
type sessionsStore struct {
	sessions.Store
	secure func() bool
}

// Get returns a cached session, setting the secure cookie option dynamically.
func (st *sessionsStore) Get(r *http.Request, name string) (s *sessions.Session, err error) {
	defer st.setSecureOptions(s)
	return st.Store.Get(r, name)
}

// New creates and returns a new session with the secure cookie setting option set
// dynamically.
func (st *sessionsStore) New(r *http.Request, name string) (s *sessions.Session, err error) {
	defer st.setSecureOptions(s)
	return st.Store.New(r, name)
}

func (st *sessionsStore) setSecureOptions(s *sessions.Session) {
	if s != nil {
		if s.Options == nil {
			s.Options = new(sessions.Options)
		}

		setSessionSecureOptions(s.Options, st.secure())
	}
}

var mockSessionStore sessions.Store

// newSessionStore creates a new session store backed by Redis.
func newSessionStore() sessions.Store {
	if mockSessionStore != nil {
		return mockSessionStore
	}

	rstore := &redistore.RediStore{
		Pool:   redispool.Store.Pool(),
		Codecs: securecookie.CodecsFromPairs([]byte(sessionCookieKey)),
		Options: &sessions.Options{
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400 * 30, // 30 days, default of the library
		},
		DefaultMaxAge: 60 * 20, // 20 minutes seems like a reasonable default
	}
	rstore.SetMaxLength(4096)
	rstore.SetSerializer(redistore.GobSerializer{})
	rstore.SetKeyPrefix("session_")

	secureCookie := func() bool {
		return conf.ExternalURLParsed().Scheme == "https"
	}

	setSessionSecureOptions(rstore.Options, secureCookie())

	return &sessionsStore{
		Store:  rstore,
		secure: secureCookie,
	}
}

// setSessionSecureOptions set the values of the session.Options.Secure
// and session.Options.SameSite fields depending on the value of the
// secure field.
func setSessionSecureOptions(opts *sessions.Options, secure bool) {
	// if Sourcegraph is running via:
	//  * HTTP:  set "SameSite=Lax" in session cookie - users can sign in, but won't be able to use the
	// 			 browser extension. Note that users will be able to use the browser extension once they
	// 			 configure their instance to use HTTPS.
	// 	* HTTPS: set "SameSite=None" in session cookie - users can sign in, and will be able to use the
	// 			 browser extension.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/6167 for more information.
	opts.SameSite = http.SameSiteLaxMode
	if secure {
		opts.SameSite = http.SameSiteNoneMode
	}

	opts.Secure = secure
}

// SetData sets the session data at the key. The session data is a map of keys to values. If no
// session exists, a new session is created.
//
// The value is JSON-encoded before being stored.
func SetData(w http.ResponseWriter, r *http.Request, key string, value any) error {
	sessionStore := newSessionStore()

	session, err := sessionStore.Get(r, cookieName)
	if err != nil {
		return errors.WithMessage(err, "getting session")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("encoding JSON session data for %q", key))
	}
	session.Values[key] = data
	if err := session.Save(r, w); err != nil {
		return errors.WithMessage(err, "saving session")
	}
	return nil
}

// GetData reads the session data at the key into the data structure addressed by value (which must
// be a pointer).
//
// The value is JSON-decoded from the raw bytes stored by the call to SetData.
func GetData(r *http.Request, key string, value any) error {
	sessionStore := newSessionStore()

	session, err := sessionStore.Get(r, cookieName)
	if err != nil {
		return errors.WithMessage(err, "getting session")
	}
	if data, ok := session.Values[key]; ok {
		if err := json.Unmarshal(data.([]byte), value); err != nil {
			return errors.WithMessage(err, fmt.Sprintf("decoding JSON session data for %q", key))
		}
	}
	return nil
}

// SetActorFromUser creates an actor from a user, sets it in the session, and
// returns a context with the user attached.
//
// 🚨 SECURITY: Should only be called after user is successfully authenticated.
func SetActorFromUser(ctx context.Context, w http.ResponseWriter, r *http.Request, user *types.User, expiryPeriod time.Duration) (context.Context, error) {
	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return ctx, err
	}

	if info.IsExpired() && !user.SiteAdmin {
		return ctx, errors.New("Sourcegraph license is expired. Only admins are allowed to sign in.")
	}

	// Authentication passed at this point, this is our actor
	act := sgactor.Actor{
		UID: user.ID,
	}

	// Add actor to the context, because we return it
	ctx = actor.WithActor(ctx, &act)

	// Write the session cookie with SetActor
	return ctx, SetActor(w, r, &act, expiryPeriod, user.CreatedAt)
}

// SetActor sets the actor in the session, or removes it if actor == nil. If no session exists, a
// new session is created.
//
// If expiryPeriod is 0, the default expiry period is used.
func SetActor(w http.ResponseWriter, r *http.Request, actor *actor.Actor, expiryPeriod time.Duration, userCreatedAt time.Time) error {
	var value *sessionInfo
	if actor != nil {
		if expiryPeriod == 0 {
			if cfgExpiry, err := time.ParseDuration(conf.Get().AuthSessionExpiry); err == nil {
				expiryPeriod = cfgExpiry
			} else { // if there is no valid session duration, fall back to the default one
				expiryPeriod = defaultExpiryPeriod
			}
		}
		removeSignOutCookieIfSet(r, w)

		value = &sessionInfo{Actor: actor, ExpiryPeriod: expiryPeriod, LastActive: time.Now(), UserCreatedAt: userCreatedAt}
	}
	return SetData(w, r, "actor", value)
}

// removeSignOutCookieIfSet removes the sign-out cookie if it is set.
func removeSignOutCookieIfSet(r *http.Request, w http.ResponseWriter) {
	if HasSignOutCookie(r) {
		http.SetCookie(w, &http.Cookie{Name: SignOutCookie, Value: "", MaxAge: -1})
	}
}

// HasSignOutCookie returns true if the given request has a sign-out cookie.
func HasSignOutCookie(r *http.Request) bool {
	ck, err := r.Cookie(SignOutCookie)
	if err != nil {
		return false
	}
	return ck != nil
}

// SetSignOutCookie sets a sign-out cookie on the given response.
func SetSignOutCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   SignOutCookie,
		Value:  "true",
		Secure: true,
		Path:   "/",
	})
}

// hasSessionCookie returns true if the given request has a session cookie.
func hasSessionCookie(r *http.Request) bool {
	c, _ := r.Cookie(cookieName)
	return c != nil
}

// deleteSession deletes the current session. If an error occurs, it returns the error but does not
// write an HTTP error response.
//
// It should only be used when there is an unrecoverable, permanent error in the session data. To
// sign out the current user, use SetActor(r, nil).
func deleteSession(w http.ResponseWriter, r *http.Request) error {
	if !hasSessionCookie(r) {
		return nil // nothing to do
	}

	sessionStore := newSessionStore()

	session, err := sessionStore.Get(r, cookieName)
	session.Options.MaxAge = -1 // expire immediately
	if err == nil {
		err = session.Save(r, w)
	}
	if err != nil && hasSessionCookie(r) {
		// Failsafe: delete the client's cookie even if the session store is unavailable.
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
	}
	return errors.WithMessage(err, "deleting session")
}

// InvalidateSessionCurrentUser invalidates all sessions for the current user.
func InvalidateSessionCurrentUser(w http.ResponseWriter, r *http.Request, db database.DB) error {
	a := actor.FromContext(r.Context())
	err := db.Users().InvalidateSessionsByID(r.Context(), a.UID)
	if err != nil {
		return err
	}

	// We make sure the session is actually removed from the client and from Redis
	// because SetData actually reuses the client session cookie if it exists.
	// See https://github.com/sourcegraph/security-issues/issues/136
	return deleteSession(w, r)
}

// InvalidateSessionsByIDs is a bulk action.
func InvalidateSessionsByIDs(ctx context.Context, db database.DB, ids []int32) error {
	return db.Users().InvalidateSessionsByIDs(ctx, ids)
}

// CookieMiddleware is an http.Handler middleware that authenticates
// future HTTP request via cookie.
func CookieMiddleware(logger log.Logger, db database.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		ctx, done := authenticateByCookie(logger, db, r, w)
		if done {
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CookieMiddlewareWithCSRFSafety is a middleware that authenticates HTTP requests using the
// provided cookie (if any), *only if* one of the following is true.
//
//   - The request originates from a trusted origin (the same origin, browser extension origin, or one
//     in the site configuration corsOrigin allow list.)
//   - The request has the special X-Requested-With header present, which is only possible to send in
//     browsers if the request passed the CORS preflight request (see the handleCORSRequest function.)
//
// If one of the above are not true, the request is still allowed to proceed but will be
// unauthenticated unless some other authentication is provided, such as an access token.
func CookieMiddlewareWithCSRFSafety(
	logger log.Logger,
	db database.DB,
	next http.Handler,
	corsAllowHeader string,
	isTrustedOrigin func(*http.Request) bool,
) http.Handler {
	corsAllowHeader = textproto.CanonicalMIMEHeaderKey(corsAllowHeader)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie, Authorization, "+corsAllowHeader)

		// Does the request have the X-Requested-With header? If so, it's trusted.
		_, isTrusted := r.Header[corsAllowHeader]
		if !isTrusted {
			// The request doesn't have the X-Requested-With header.
			// Did the request come from a trusted origin? If so, it's trusted.
			isTrusted = isTrustedOrigin(r)
		}
		if isTrusted {
			ctx, done := authenticateByCookie(logger, db, r, w)
			if done {
				return
			}
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func authenticateByCookie(logger log.Logger, db database.DB, r *http.Request, w http.ResponseWriter) (_ context.Context, done bool) {
	span, ctx := trace.New(r.Context(), "session.authenticateByCookie")
	defer span.End()
	logger = trace.Logger(ctx, logger)

	// If the request is already authenticated from a cookie (and not a token), then do not clobber the request's existing
	// authenticated actor with the actor (if any) derived from the session cookie.
	if a := actor.FromContext(ctx); a.IsAuthenticated() && a.FromSessionCookie {
		span.SetAttributes(
			attribute.Bool("authenticated", true),
			attribute.Bool("fromSessionCookie", true),
		)
		if hasSessionCookie(r) {
			// Delete the session cookie to avoid confusion. This occurs most often when
			// switching the auth provider to http-header; in that case, we want to rely on
			// the http-header auth provider for auth, not the user's old session.
			span.AddEvent("has session cookie, deleting session")
			_ = deleteSession(w, r)
		}
		return ctx, false // unchanged
	}

	var info *sessionInfo
	if err := GetData(r, "actor", &info); err != nil {
		if errors.HasType[*net.OpError](err) {
			// If fetching session info failed because of a Redis error, return empty Context
			// without deleting the session cookie and throw an internal server error.
			// This prevents background requests made by off-screen tabs from signing
			// the user out during a server update.
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "failed to read session actor")
			span.AddEvent("redis connection refused")
			return ctx, true
		}

		if !strings.Contains(err.Error(), "illegal base64 data at input byte 36") {
			// Skip log if the error message indicates the cookie value was a JWT (which almost
			// certainly means that the cookie was a pre-2.8 SAML cookie, so this error will only
			// occur once and the user will be automatically redirected to the SAML auth flow).
			logger.Warn("error reading session actor - the session cookie was invalid and will be cleared (this error can be safely ignored unless it persists)",
				log.Error(err))
		}
		_ = deleteSession(w, r) // clear the bad value
		span.SetError(err)
		return ctx, false
	}
	if info != nil {
		logger := logger.With(log.Int32("uid", info.Actor.UID))
		span.SetAttributes(attribute.String("uid", info.Actor.UIDString()))

		// Check expiry
		if info.LastActive.Add(info.ExpiryPeriod).Before(time.Now()) {
			_ = deleteSession(w, r) // clear the bad value
			return actor.WithActor(ctx, &actor.Actor{}), false
		}

		// Check that user still exists.
		usr, err := db.Users().GetByID(ctx, info.Actor.UID)
		if err != nil {
			if errcode.IsNotFound(err) {
				_ = deleteSession(w, r) // clear the bad value
			} else {
				// Don't delete session, since the error might be an ephemeral DB error, and we don't
				// want that to cause all active users to be signed out.
				logger.Error("error looking up user for session", log.Error(err))
			}
			span.SetError(err)
			return ctx, false // not authenticated
		}

		licenseInfo, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			return ctx, false
		}

		if licenseInfo.IsExpired() && !usr.SiteAdmin {
			_ = deleteSession(w, r) // Delete session since only admins are allowed
			return ctx, false
		}

		// Check that the session is still valid
		if info.LastActive.Before(usr.InvalidatedSessionsAt) {
			span.SetAttributes(attribute.Bool("expired", true))
			_ = deleteSession(w, r) // Delete the now invalid session
			return ctx, false
		}

		// If the session does not have the user's creation date, it's an old (valid)
		// session from before the check was introduced. In that case, we manually
		// set the user creation date
		if info.UserCreatedAt.IsZero() {
			info.UserCreatedAt = usr.CreatedAt
			if err := SetData(w, r, "actor", info); err != nil {
				logger.Error("error setting user creation timestamp", log.Error(err))
				return ctx, false
			}
		}

		// Verify that the user's creation date in the database matches what is stored
		// in the session. If not, invalidate the session immediately.
		if !info.UserCreatedAt.Equal(usr.CreatedAt) {
			span.SetError(errors.New("user creation date does not match database"))
			_ = deleteSession(w, r)
			return ctx, false
		}

		// Renew session
		if time.Since(info.LastActive) > 5*time.Minute {
			info.LastActive = time.Now()
			if err := SetData(w, r, "actor", info); err != nil {
				logger.Error("error renewing session", log.Error(err))
				return ctx, false
			}
		}

		span.SetAttributes(attribute.Bool("authenticated", true))
		info.Actor.FromSessionCookie = true
		return actor.WithActor(ctx, info.Actor), false
	}

	return ctx, false
}
