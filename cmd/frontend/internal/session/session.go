package session

import (
	"context"
	"encoding/json"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

// actorSessionStore stores the actor-based user session.
var actorSessionStore *Store

type sessionInfo struct {
	Actor  *actor.Actor `json:"actor"`
	Expiry time.Time    `json:"expiry"`
}

// InitSessionStore initializes the session store.
func InitSessionStore(secureCookie bool) {
	var err error
	actorSessionStore, err = NewStore("sg-session", "actor", secureCookie, nil)
	if err != nil {
		panic(err)
	}
}

// StartNewSession starts a new session with authentication for the given uid and a given expiration time.
func StartNewSession(w http.ResponseWriter, r *http.Request, actor *actor.Actor, expiry time.Time) error {
	sessionJSON, err := json.Marshal(sessionInfo{Actor: actor, Expiry: expiry})
	if err != nil {
		return err
	}
	actorSessionStore.StartNewSession(w, r, sessionJSON)
	return nil
}

// DeleteSession deletes the current session.
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	actorSessionStore.DeleteSession(w, r)
}

// SessionCookie returns the session cookie from the header of the given request.
func SessionCookie(r *http.Request) string {
	return actorSessionStore.Cookie(r)
}

// CookieMiddleware is an http.Handler middleware that authenticates
// future HTTP request via cookie.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(authenticateByCookie(r, w)))
	})
}

// CookieOrSessionMiddleware is like CookieMiddleware, but also inspects the HTTP Authorization
// header for a session cookie and uses that if it exists.
func CookieOrSessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "session" {
			next.ServeHTTP(w, r.WithContext(AuthenticateBySession(r.Context(), parts[1])))
		} else {
			next.ServeHTTP(w, r.WithContext(authenticateByCookie(r, w)))
		}
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
			r = r.WithContext(authenticateByCookie(r, w))
		}
		next.ServeHTTP(w, r)
	})
}

// AuthenticateBySession authenticates the context with the given session cookie.
func AuthenticateBySession(ctx context.Context, sessionCookie string) context.Context {
	fakeRequest := &http.Request{Header: http.Header{"Cookie": []string{actorSessionStore.name + "=" + sessionCookie}}}
	return authenticateByCookie(fakeRequest.WithContext(ctx), nil)
}

// authenticateByCookie returns an authenticated Context using the session cookie in a request.
// If the session is expired, we delete the session and the cookie.
func authenticateByCookie(r *http.Request, w http.ResponseWriter) context.Context {
	sessionJSON, err := actorSessionStore.GetSession(r)
	if err != nil {
		log15.Error("error getting session", "error", err)
		return r.Context()
	}

	if sessionJSON == nil {
		return r.Context()
	}

	var info sessionInfo
	if err := json.Unmarshal(sessionJSON, &info); err != nil {
		log15.Error("error unmarshalling session", "error", err)
		return r.Context()
	}

	if !info.Expiry.After(time.Now()) {
		if w != nil {
			actorSessionStore.DeleteSession(w, r)
		}
		return r.Context()
	}

	return actor.WithActor(r.Context(), info.Actor)
}
