package session

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/textproto"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
)

var sessionStore sessions.Store
var sessionStoreRedis = env.Get("SRC_SESSION_STORE_REDIS", "redis-store:6379", "redis used for storing sessions")
var sessionCookieKey = env.Get("SRC_SESSION_COOKIE_KEY", "", "secret key used for securing the session cookies")

// DefaultExpiryPeriod is the default session expiry period (if none is specified explicitly): 90 days.
const DefaultExpiryPeriod = 90 * 24 * time.Hour

// sessionInfo is the information we store in the session. The gorilla/sessions library doesn't appear to
// enforce the maxAge field in its session store implementations, so we include the expiry here.
type sessionInfo struct {
	Actor        *actor.Actor  `json:"actor"`
	LastActive   time.Time     `json:"lastActive"`
	ExpiryPeriod time.Duration `json:"expiryPeriod"`

	// DEPRECATED. Can be removed after December 31, 2017
	Expiry time.Time `json:"expiry"`
}

// SetSessionStore sets the backing store used for storing sessions on the server. It should be called exactly once.
func SetSessionStore(s sessions.Store) {
	sessionStore = s
}

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

	DeleteSession(w, r)

	session, err := sessionStore.New(&http.Request{}, "sg-session") // workaround: not passing the request forces a new session
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

// DeleteSession deletes the current session.
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "sg-session")
	if err != nil {
		// See the other "conf.AuthSAML() == nil" line below for why it's OK to skip logging when using SAML.
		if conf.AuthSAML() == nil {
			log15.Error("error getting session", "error", err)
		}
	}
	session.Options.MaxAge = -1
	if err = session.Save(r, w); err != nil {
		log15.Error("error saving session", "error", err)
	}
}

// CookieMiddleware is an http.Handler middleware that authenticates
// future HTTP request via cookie.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie")
		next.ServeHTTP(w, r.WithContext(authenticateByCookie(r, w)))
	})
}

// CookieMiddlewareIfHeader is an http.Handler middleware that
// authenticates future HTTP requests via cookie, *only if* a specific
// header is present. Typically X-Requested-By is used as the header
// name. This protects against CSRF (see
// https://security.stackexchange.com/questions/23371/csrf-protection-with-custom-headers-and-without-validating-token).
func CookieMiddlewareIfHeader(next http.Handler, headerName string) http.Handler {
	headerName = textproto.CanonicalMIMEHeaderKey(headerName)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Cookie, Authorization, "+headerName)
		if _, ok := r.Header[headerName]; ok {
			r = r.WithContext(authenticateByCookie(r, w))
		}
		next.ServeHTTP(w, r)
	})
}

func authenticateByCookie(r *http.Request, w http.ResponseWriter) context.Context {
	session, err := sessionStore.Get(r, "sg-session")
	if err != nil {
		// Ignore this error (and skip logging) when using SAML because SAML's cookies have the same
		// name (sg-session) but are actually SAML-specific JSON Web Tokens (JWTs) that are not
		// validated using our own session store.
		if conf.AuthSAML() == nil {
			log15.Error("error getting session", "error", err)
		}
		return r.Context()
	}

	if actorJSON, ok := session.Values["actor"]; ok {
		var info sessionInfo
		if err := json.Unmarshal(actorJSON.([]byte), &info); err != nil {
			log15.Error("error unmarshalling actor", "error", err)
			DeleteSession(w, r) // so that we clear the bad value
			return r.Context()
		}

		// Check expiry
		if info.LastActive.Add(info.ExpiryPeriod).Before(time.Now()) {
			DeleteSession(w, r)
			return actor.WithActor(r.Context(), &actor.Actor{})
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
