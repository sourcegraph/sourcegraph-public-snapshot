package session

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/boj/redistore"
	"github.com/gorilla/sessions"
)

var (
	sessionStoreRedis = env.Get("SRC_SESSION_STORE_REDIS", "redis-store:6379", "redis used for storing sessions")
	sessionCookieKey  = env.Get("SRC_SESSION_COOKIE_KEY", "", "secret key used for securing the session cookies")
)

func init() {
	if sessionStoreRedis == "" {
		sessionStoreRedis = ":6379"
	}
}

// Store is Redis-backed store for sessions
type Store struct {
	Name      string
	CookieKey string
	store     sessions.Store
}

// Init initializes the session store.
func NewStore(name string, cookieKey string, secureCookie bool, underlyingStore sessions.Store) (*Store, error) {
	if underlyingStore == nil {
		redisStore, err := redistore.NewRediStore(10, "tcp", sessionStoreRedis, "", []byte(sessionCookieKey))
		if err != nil {
			return nil, err
		}
		redisStore.Options.Path = "/"
		redisStore.Options.HttpOnly = true
		redisStore.Options.Secure = secureCookie
		underlyingStore = redisStore
	}
	return &Store{
		Name:      name,
		CookieKey: cookieKey,
		store:     underlyingStore,
	}, nil
}

// StartNewSession starts a new session with authentication for the given uid.
func (s *Store) StartNewSession(w http.ResponseWriter, r *http.Request, val []byte) error {
	s.DeleteSession(w, r)

	session, err := s.store.New(&http.Request{}, s.Name) // workaround: not passing the request forces a new session
	if err != nil {
		log15.Error("error creating session", "error", err)
	}
	session.Values[s.CookieKey] = val
	if err = session.Save(r, w); err != nil {
		log15.Error("error saving session", "error", err)
	}

	return nil
}

// GetSession gets the current session token value, nil if it doesn't exist
func (s *Store) GetSession(r *http.Request) ([]byte, error) {
	session, err := s.store.Get(r, s.Name)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	val := session.Values[s.CookieKey]
	if val == nil {
		return nil, nil
	} else if val, ok := val.([]byte); ok {
		return val, nil
	}
	return nil, fmt.Errorf("invalid value for session: %v", val)
}

// DeleteSession deletes the current session.
func (s *Store) DeleteSession(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.Get(r, s.Name)
	if err != nil {
		log15.Error("error getting session", "error", err)
	}
	session.Options.MaxAge = -1
	if err = session.Save(r, w); err != nil {
		log15.Error("error saving session", "error", err)
	}
}

// SessionCookie returns the session cookie from the header of the given request.
func (s *Store) Cookie(r *http.Request) string {
	c, err := r.Cookie(s.CookieKey)
	if err != nil {
		return ""
	}
	return c.Value
}
