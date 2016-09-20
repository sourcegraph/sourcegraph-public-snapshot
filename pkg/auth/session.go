package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/boj/redistore"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

var sessionStore *redistore.RediStore

// InitSessionStore initializes the session store.
func InitSessionStore(secureCookie bool) {
	sessionStoreRedis := os.Getenv("SRC_SESSION_STORE_REDIS")
	if sessionStoreRedis == "" {
		sessionStoreRedis = ":6379"
	}
	var err error
	sessionStore, err = redistore.NewRediStore(10, "tcp", sessionStoreRedis, "", []byte(os.Getenv("SRC_SESSION_COOKIE_KEY")))
	if err != nil {
		panic(err)
	}
	sessionStore.Options.Path = "/"
	sessionStore.Options.HttpOnly = true
	sessionStore.Options.Secure = secureCookie
}

// StartNewSession starts a new session with authentication for the given uid.
func StartNewSession(w http.ResponseWriter, r *http.Request, actor *Actor) error {
	DeleteSession(w, r)

	session, err := sessionStore.New(&http.Request{}, "sg-session") // workaround: not passing the request forces a new session
	if err != nil {
		log15.Error("error creating session", "error", err)
	}
	actorJSON, err := json.Marshal(actor)
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

// CookieMiddleware is an http.Handler middleware that authenticates
// future HTTP request via cookie.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(authenticateByCookie(r)))
	})
}

// SessionCookie returns the session cookie from the header of the given request.
func SessionCookie(r *http.Request) string {
	c, err := r.Cookie("sg-session")
	if err != nil {
		return ""
	}
	return c.Value
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
		var actor Actor
		if err := json.Unmarshal(actorJSON.([]byte), &actor); err != nil {
			log15.Error("error unmarshalling actor", "error", err)
			return r.Context()
		}

		token, err := NewAccessToken(&actor, nil, 7*24*time.Hour)
		if err != nil {
			log15.Error("error creating access token", "error", err)
			return r.Context()
		}

		return WithActor(sourcegraph.WithAccessToken(r.Context(), token), &actor)
	}

	return r.Context()
}
