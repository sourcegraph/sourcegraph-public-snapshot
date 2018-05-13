package session

import (
	"context"
	"encoding/json"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
)

var sessionStore sessions.Store
var sessionStoreRedis = env.Get("SRC_SESSION_STORE_REDIS", "redis-store:6379", "redis used for storing sessions")
var sessionCookieKey = env.Get("SRC_SESSION_COOKIE_KEY", "", "secret key used for securing the session cookies")

// DefaultExpiryPeriod is the default session expiry period (if none is specified explicitly): 90 days.
const DefaultExpiryPeriod = 90 * 24 * time.Hour

// cookieName is the name of the HTTP cookie that stores the session ID.
const cookieName = "sg-session"

// sessionInfo is the information we store in the session. The gorilla/sessions library doesn't appear to
// enforce the maxAge field in its session store implementations, so we include the expiry here.
type sessionInfo struct {
	Actor        *actor.Actor  `json:"actor"`
	LastActive   time.Time     `json:"lastActive"`
	ExpiryPeriod time.Duration `json:"expiryPeriod"`
}

// SetSessionStore sets the backing store used for storing sessions on the server. It should be called exactly once.
func SetSessionStore(s sessions.Store) {
	sessionStore = s
}

// NewRedisStore creates a new session store backed by Redis.
func NewRedisStore(secureCookie bool) sessions.Store {
	if sessionStoreRedis == "" {
		sessionStoreRedis = ":6379"
	}
	rstore, err := redistore.NewRediStore(10, "tcp", sessionStoreRedis, "", []byte(sessionCookieKey))
	if err != nil {
		waitForRedis(rstore)
	}
	rstore.Options.Path = "/"
	rstore.Options.HttpOnly = true
	rstore.Options.Secure = secureCookie
	return rstore
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

// StartNewSession starts a new session with authentication for the given uid. If expiryPeriod is zero
// the defaultExpiryPeriod value is used.
func StartNewSession(w http.ResponseWriter, r *http.Request, actor *actor.Actor, expiryPeriod time.Duration) error {
	if expiryPeriod == 0 {
		expiryPeriod = DefaultExpiryPeriod
	}

	if err := DeleteSession(w, r); err != nil {
		log15.Error("Error deleting previous session when starting new session.", "err", err)
	}

	session, err := sessionStore.New(&http.Request{}, cookieName) // workaround: not passing the request forces a new session
	if err != nil {
		log15.Error("error creating session", "error", err)
	}
	actorJSON, err := json.Marshal(sessionInfo{Actor: actor, ExpiryPeriod: expiryPeriod, LastActive: time.Now()})
	if err != nil {
		return err
	}
	session.Values["actor"] = actorJSON
	if err = session.Save(r, w); err != nil {
		log15.Error("error saving session", "error", err)
	}

	return nil
}

// ignoreSessionCookieError reports whether session cookie errors should be ignored and not
// logged. It is true iff the auth provider is SAML because SAML's cookies have the same name
// (sg-session) but are actually SAML-specific JSON Web Tokens (JWTs) that are not validated using
// our own session store. Therefore they always produce an error.
//
// TODO(sqs): Make it so that our SAML cookies use a different name (and do this without logging
// all SAML users out).
func ignoreSessionCookieError() bool {
	return conf.AuthProvider().Saml != nil
}

func hasSessionCookie(r *http.Request) bool {
	c, _ := r.Cookie(cookieName)
	return c != nil
}

// DeleteSession deletes the current session. If an error occurs, it returns the error but does not
// write an HTTP error response.
func DeleteSession(w http.ResponseWriter, r *http.Request) error {
	if !hasSessionCookie(r) {
		return nil // nothing to do
	}

	session, err := sessionStore.Get(r, cookieName)
	if err == nil {
		session.Options.MaxAge = -1 // expire immediately
		err = session.Save(r, w)
	}
	if err != nil && hasSessionCookie(r) {
		// Failsafe: delete the client's cookie even if the session store is unavailable.
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
	}
	return errors.WithMessage(err, "deleting session")
}

// CookieMiddleware is an http.Handler middleware that authenticates
// future HTTP request via cookie.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		next.ServeHTTP(w, r.WithContext(authenticateByCookie(r, w)))
	})
}

// CookieMiddlewareWithCSRFSafety is a middleware that authenticates HTTP requests using the
// provided cookie (if any), *only if* the request is a non-simple CORS request (see
// https://www.w3.org/TR/cors/#cross-origin-request-with-preflight-0). This relies on the client's
// CORS checks to guarantee that one of the following is true, thereby protecting against CSRF
// attacks:
//
// - The request originates from the same origin. -OR-
//
// - The request is cross-origin but passed the CORS preflight check (because otherwise the
//   preflight OPTIONS reponse from secureHeadersMiddleware would have caused the browser to refuse
//   to send this HTTP request).
//
// To determine if it's a non-simple CORS request, it checks for the presence of either
// "Content-Type: application/json; charset=utf-8" or a non-empty HTTP request header whose name is
// given in corsAllowHeader.
//
// NOTE: As a special temporary case, if the request path begins with /.api/telemetry/log/, it uses
// cookies for authentication. See https://github.com/sourcegraph/sourcegraph/issues/10901 for why.
//
// If the request is a simple CORS request, or if neither of these is true, then the cookie is not
// used to authenticate the request. The request is still allowed to proceed (but will be
// unauthenticated unless some other authentication is provided, such as an access token).
func CookieMiddlewareWithCSRFSafety(next http.Handler, corsAllowHeader string, isTrustedOrigin func(*http.Request) bool) http.Handler {
	corsAllowHeader = textproto.CanonicalMIMEHeaderKey(corsAllowHeader)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie, Authorization, "+corsAllowHeader)

		_, isTrusted := r.Header[corsAllowHeader]
		if !isTrusted {
			isTrusted = isTrustedOrigin(r)
		}
		if !isTrusted {
			contentType := r.Header.Get("Content-Type")
			isTrusted = contentType == "application/json" || contentType == "application/json; charset=utf-8"
		}
		if !isTrusted {
			// See NOTE in docstring for why this is special-case allowed.
			isTrusted = strings.HasPrefix(r.URL.Path, "/.api/telemetry/log/")
		}
		if isTrusted {
			r = r.WithContext(authenticateByCookie(r, w))
		}

		next.ServeHTTP(w, r)
	})
}

func authenticateByCookie(r *http.Request, w http.ResponseWriter) context.Context {
	// If the request is already authenticated, then do not clobber the request's existing
	// authenticated actor with the actor (if any) derived from the session cookie.
	if actor.FromContext(r.Context()).IsAuthenticated() {
		if hasSessionCookie(r) {
			// Delete the session cookie to avoid confusion. (This occurs most often when switching
			// the auth provider to http-header; in that case, we want to rely on the http-header
			// auth provider for auth, not the user's old session.
			_ = DeleteSession(w, r)
		}
		return r.Context() // unchanged
	}

	session, err := sessionStore.Get(r, cookieName)
	if err != nil {
		if !ignoreSessionCookieError() {
			log15.Error("error getting session", "error", err)
		}
		return r.Context()
	}

	if actorJSON, ok := session.Values["actor"]; ok {
		var info sessionInfo
		if err := json.Unmarshal(actorJSON.([]byte), &info); err != nil {
			log15.Error("error unmarshalling actor", "error", err)
			_ = DeleteSession(w, r) // clear the bad value
			return r.Context()
		}

		// Check expiry
		if info.LastActive.Add(info.ExpiryPeriod).Before(time.Now()) {
			_ = DeleteSession(w, r) // clear the bad value
			return actor.WithActor(r.Context(), &actor.Actor{})
		}

		// Check that user still exists.
		if _, err := db.Users.GetByID(r.Context(), info.Actor.UID); err != nil {
			if errcode.IsNotFound(err) {
				_ = DeleteSession(w, r) // clear the bad value
			} else {
				// Don't delete session, since the error might be an ephemeral DB error, and we don't
				// want that to cause all active users to be signed out.
				log15.Error("Error looking up user for session.", "uid", info.Actor.UID, "error", err)
			}
			return r.Context() // not authenticated
		}

		// Renew session
		if time.Now().Sub(info.LastActive) > 5*time.Minute {
			info.LastActive = time.Now()
			newActorJSON, err := json.Marshal(info)
			if err != nil {
				log15.Error("error renewing session", "error", err)
				return r.Context()
			}
			session.Values["actor"] = newActorJSON
			if err := session.Save(r, w); err != nil {
				log15.Error("error saving session", "error", err)
				return r.Context()
			}
		}

		return actor.WithActor(r.Context(), info.Actor)
	}

	return r.Context()
}
