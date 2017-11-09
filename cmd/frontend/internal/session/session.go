package session

import (
	"context"
	"encoding/json"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
)

var sessionStore sessions.Store
var sessionStoreRedis = env.Get("SRC_SESSION_STORE_REDIS", "redis-store:6379", "redis used for storing sessions")
var sessionCookieKey = env.Get("SRC_SESSION_COOKIE_KEY", "", "secret key used for securing the session cookies")

// sessionInfo is the information we store in the session. The gorilla/sessions library doesn't appear to
// enforce the maxAge field in its session store implementations, so we include the expiry here.
type sessionInfo struct {
	Actor  *actor.Actor `json:"actor"`
	Expiry time.Time    `json:"expiry"`
}

// SetSessionStore sets the backing store used for storing sessions on the server. It should be called exactly once.
func SetSessionStore(s sessions.Store) {
	sessionStore = s
}

func MustNewRedisStore(secureCookie bool) sessions.Store {
	if sessionStoreRedis == "" {
		sessionStoreRedis = ":6379"
	}
	rstore, err := redistore.NewRediStore(10, "tcp", sessionStoreRedis, "", []byte(sessionCookieKey))
	if err != nil {
		panic(err)
	}
	rstore.Options.Path = "/"
	rstore.Options.HttpOnly = true
	rstore.Options.Secure = secureCookie
	return rstore
}

// StartNewSession starts a new session with authentication for the given uid.
func StartNewSession(w http.ResponseWriter, r *http.Request, actor *actor.Actor, expiry time.Time) error {
	DeleteSession(w, r)

	session, err := sessionStore.New(&http.Request{}, "sg-session") // workaround: not passing the request forces a new session
	if err != nil {
		log15.Error("error creating session", "error", err)
	}
	actorJSON, err := json.Marshal(sessionInfo{Actor: actor, Expiry: expiry})
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
		log15.Error("error getting session", "error", err)
	}
	session.Options.MaxAge = -1
	if err = session.Save(r, w); err != nil {
		log15.Error("error saving session", "error", err)
	}
}

// CookieOrSessionMiddleware is like CookieMiddleware, but also inspects the HTTP Authorization
// header for a session cookie and uses that if it exists.
func CookieOrSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "session" {
			next.ServeHTTP(w, r.WithContext(AuthenticateBySession(r.Context(), parts[1])))
		} else {
			next.ServeHTTP(w, r.WithContext(authenticateByCookie(r)))
		}
	})
}

// CookieMiddleware is an http.Handler middleware that authenticates
// future HTTP request via cookie.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(authenticateByCookie(r)))
	})
}

// CookieMiddlewareIfHeader is an http.Handler middleware that
// authenticates future HTTP requests via cookie, *only if* a specific
// header is present. Typically X-Requested-By is used as the header
// name. This protects against CSRF (see
// https://security.stackexchange.com/questions/23371/csrf-protection-with-custom-headers-and-without-validating-token).
func CookieMiddlewareIfHeader(next http.Handler, headerName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Header[textproto.CanonicalMIMEHeaderKey(headerName)]; ok {
			r = r.WithContext(authenticateByCookie(r))
		}
		next.ServeHTTP(w, r)
	})
}

// AuthenticateBySession authenticates the context with the given session cookie.
func AuthenticateBySession(ctx context.Context, sessionCookie string) context.Context {
	fakeRequest := &http.Request{Header: http.Header{"Cookie": []string{"sg-session=" + sessionCookie}}}
	return authenticateByCookie(fakeRequest.WithContext(ctx))
}

func authenticateByCookie(r *http.Request) context.Context {
	session, err := sessionStore.Get(r, "sg-session")
	if err != nil {
		log15.Error("error getting session", "error", err)
		return r.Context()
	}

	if actorJSON, ok := session.Values["actor"]; ok {
		var info sessionInfo
		if err := json.Unmarshal(actorJSON.([]byte), &info); err != nil {
			log15.Error("error unmarshalling actor", "error", err)
			return r.Context()
		}

		if info.Actor == nil || info.Expiry.Before(time.Now()) {
			return actor.WithActor(r.Context(), &actor.Actor{})
		}
		return actor.WithActor(r.Context(), info.Actor)
	}

	return r.Context()
}

// SessionCookie returns the session cookie from the header of the given request.
func SessionCookie(r *http.Request) string {
	c, err := r.Cookie("sg-session")
	if err != nil {
		return ""
	}
	return c.Value
}
