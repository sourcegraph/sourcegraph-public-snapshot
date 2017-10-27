package session

import (
	"context"
	"encoding/json"
	"net/http"
	"net/textproto"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

var actorSessionStore *Store

// InitSessionStore initializes the session store.
func InitSessionStore(secureCookie bool) {
	var err error
	actorSessionStore, err = NewStore("sg-session", "actor", secureCookie, nil)
	if err != nil {
		panic(err)
	}
}

// StartNewSession starts a new session with authentication for the given uid.
func StartNewSession(w http.ResponseWriter, r *http.Request, actor *actor.Actor) error {

	actorJSON, err := json.Marshal(actor)
	if err != nil {
		return err
	}
	actorSessionStore.StartNewSession(w, r, actorJSON)
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
	fakeRequest := &http.Request{Header: http.Header{"Cookie": []string{actorSessionStore.Name + "=" + sessionCookie}}}
	return authenticateByCookie(fakeRequest.WithContext(ctx))
}

func authenticateByCookie(r *http.Request) context.Context {
	actorJSON, err := actorSessionStore.GetSession(r)
	if err != nil {
		log15.Error("error getting session", "error", err)
		// 🚨 SECURITY: erase any existing actor
		return actor.WithActor(r.Context(), &actor.Actor{})
	}

	var a actor.Actor
	if err := json.Unmarshal(actorJSON, &a); err != nil {
		log15.Error("error unmarshalling actor", "error", err)
		// 🚨 SECURITY: erase any existing actor
		return actor.WithActor(r.Context(), &actor.Actor{})
	}

	return actor.WithActor(r.Context(), &a)
}
