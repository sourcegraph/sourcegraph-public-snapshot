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

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/inconshreveable/log15"

	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
)

const SignOutCookie = "sg-signout"

var (
	sessionStore     sessions.Store
	sessionCookieKey = env.Get("SRC_SESSION_COOKIE_KEY", "", "secret key used for securing the session cookies")
)

// defaultExpiryPeriod is the default session expiry period (if none is specified explicitly): 90 days.
const defaultExpiryPeriod = 90 * 24 * time.Hour

// cookieName is the name of the HTTP cookie that stores the session ID.
const cookieName = "sgs"

func init() {
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		if c.SiteConfig().AuthSessionExpiry == "" {
			return nil
		}

		d, err := time.ParseDuration(c.SiteConfig().AuthSessionExpiry)
		if err != nil {
			return conf.NewSiteProblems("auth.sessionExpiry does not conform to the Go time.Duration format (https://golang.org/pkg/time/#ParseDuration). The default of 90 days will be used.")
		}
		if d == 0 {
			return conf.NewSiteProblems("auth.sessionExpiry should be greater than zero. The default of 90 days will be used.")
		}
		return nil
	})
}

// sessionInfo is the information we store in the session. The gorilla/sessions library doesn't appear to
// enforce the maxAge field in its session store implementations, so we include the expiry here.
type sessionInfo struct {
	Actor         *actor.Actor  `json:"actor"`
	LastActive    time.Time     `json:"lastActive"`
	ExpiryPeriod  time.Duration `json:"expiryPeriod"`
	UserCreatedAt time.Time     `json:"userCreatedAt"`
}

// SetSessionStore sets the backing store used for storing sessions on the server. It should be called exactly once.
func SetSessionStore(s sessions.Store) {
	sessionStore = s
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

// NewRedisStore creates a new session store backed by Redis.
func NewRedisStore(secureCookie func() bool) sessions.Store {
	var store sessions.Store
	var options *sessions.Options

	pool := redispool.Store.Pool()
	rstore, err := redistore.NewRediStoreWithPool(pool, []byte(sessionCookieKey))
	if err != nil {
		waitForRedis(rstore)
	}
	store = rstore
	options = rstore.Options

	options.Path = "/"
	options.HttpOnly = true

	setSessionSecureOptions(options, secureCookie())
	return &sessionsStore{
		Store:  store,
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

// Ping attempts to contact Redis and returns a non-nil error upon failure. It is intended to be
// used by health checks.
func Ping() error {
	if sessionStore == nil {
		return errors.New("redis session store is not available")
	}
	rstore, ok := sessionStore.(*redistore.RediStore)
	if !ok {
		// Only try to ping Redis session stores. If we add other types of session stores, add ways
		// to ping them here.
		return nil
	}
	return ping(rstore)
}

func ping(s *redistore.RediStore) error {
	conn := s.Pool.Get()
	defer conn.Close()
	data, err := conn.Do("PING")
	if err != nil {
		return err
	}
	if data != "PONG" {
		return errors.New("no pong received")
	}
	return nil
}

// waitForRedis waits up to a certain timeout for Redis to become reachable, to reduce the
// likelihood of the HTTP handlers starting to serve requests while Redis (and therefore session
// data) is still unavailable. After the timeout has elapsed, if Redis is still unreachable, it
// continues anyway (because that's probably better than the site not coming up at all).
func waitForRedis(s *redistore.RediStore) {
	const timeout = 5 * time.Second
	deadline := time.Now().Add(timeout)
	var err error
	for {
		time.Sleep(150 * time.Millisecond)
		err = ping(s)
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			log15.Warn("Redis (used for session store) failed to become reachable. Will continue trying to establish connection in background.", "timeout", timeout, "error", err)
			return
		}
	}
}

// SetData sets the session data at the key. The session data is a map of keys to values. If no
// session exists, a new session is created.
//
// The value is JSON-encoded before being stored.
func SetData(w http.ResponseWriter, r *http.Request, key string, value any) error {
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

func SetActorFromUser(ctx context.Context, w http.ResponseWriter, r *http.Request, user *types.User, expiryPeriod time.Duration) (context.Context, error) {
	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return ctx, err
	}

	if info.IsExpired() && !user.SiteAdmin {
		return ctx, errors.New("Sourcegraph license is expired. Only admins are allowed to sign in.")
	}

	// Write the session cookie
	actor := sgactor.Actor{
		UID: user.ID,
	}

	return ctx, SetActor(w, r, &actor, expiryPeriod, user.CreatedAt)
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
		RemoveSignOutCookieIfSet(r, w)

		value = &sessionInfo{Actor: actor, ExpiryPeriod: expiryPeriod, LastActive: time.Now(), UserCreatedAt: userCreatedAt}
	}
	return SetData(w, r, "actor", value)
}

// RemoveSignOutCookieIfSet removes the sign-out cookie if it is set.
func RemoveSignOutCookieIfSet(r *http.Request, w http.ResponseWriter) {
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
		next.ServeHTTP(w, r.WithContext(authenticateByCookie(logger, db, r, w)))
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
			r = r.WithContext(authenticateByCookie(logger, db, r, w))
		}

		next.ServeHTTP(w, r)
	})
}

func authenticateByCookie(logger log.Logger, db database.DB, r *http.Request, w http.ResponseWriter) context.Context {
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
		return ctx // unchanged
	}

	var info *sessionInfo
	if err := GetData(r, "actor", &info); err != nil {
		if errors.HasType(err, &net.OpError{}) {
			// If fetching session info failed because of a Redis error, return empty Context
			// without deleting the session cookie and throw an internal server error.
			// This prevents background requests made by off-screen tabs from signing
			// the user out during a server update.
			w.WriteHeader(http.StatusInternalServerError)
			span.AddEvent("redis connection refused")
			return ctx
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
		return ctx
	}
	if info != nil {
		logger := logger.With(log.Int32("uid", info.Actor.UID))
		span.SetAttributes(attribute.String("uid", info.Actor.UIDString()))

		// Check expiry
		if info.LastActive.Add(info.ExpiryPeriod).Before(time.Now()) {
			_ = deleteSession(w, r) // clear the bad value
			return actor.WithActor(ctx, &actor.Actor{})
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
			return ctx // not authenticated
		}

		licenseInfo, err := licensing.GetConfiguredProductLicenseInfo()
		if err != nil {
			return ctx
		}

		if licenseInfo.IsExpired() && !usr.SiteAdmin {
			_ = deleteSession(w, r) // Delete session since only admins are allowed
			return ctx
		}

		// Check that the session is still valid
		if info.LastActive.Before(usr.InvalidatedSessionsAt) {
			span.SetAttributes(attribute.Bool("expired", true))
			_ = deleteSession(w, r) // Delete the now invalid session
			return ctx
		}

		// If the session does not have the user's creation date, it's an old (valid)
		// session from before the check was introduced. In that case, we manually
		// set the user creation date
		if info.UserCreatedAt.IsZero() {
			info.UserCreatedAt = usr.CreatedAt
			if err := SetData(w, r, "actor", info); err != nil {
				logger.Error("error setting user creation timestamp", log.Error(err))
				return ctx
			}
		}

		// Verify that the user's creation date in the database matches what is stored
		// in the session. If not, invalidate the session immediately.
		if !info.UserCreatedAt.Equal(usr.CreatedAt) {
			span.SetError(errors.New("user creation date does not match database"))
			_ = deleteSession(w, r)
			return ctx
		}

		// Renew session
		if time.Since(info.LastActive) > 5*time.Minute {
			info.LastActive = time.Now()
			if err := SetData(w, r, "actor", info); err != nil {
				logger.Error("error renewing session", log.Error(err))
				return ctx
			}
		}

		span.SetAttributes(attribute.Bool("authenticated", true))
		info.Actor.FromSessionCookie = true
		return actor.WithActor(ctx, info.Actor)
	}

	return ctx
}
