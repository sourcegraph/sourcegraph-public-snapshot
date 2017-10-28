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
	sessionStoreRedis     = env.Get("SRC_SESSION_STORE_REDIS", "redis-store:6379", "redis used for storing sessions")
	sessionCookieKeyRedis = env.Get("SRC_SESSION_COOKIE_KEY", "", "secret key used for securing the session cookies")
)

func init() {
	if sessionStoreRedis == "" {
		sessionStoreRedis = ":6379"
	}
}

// Store is a wrapper that stores and fetches data from some underlying session store
type Store struct {
	// name is the name of the session and session cookie
	name string

	// sessionKey is the key in the underlying gorilla/session from which we fetch/set the session data.
	sessionKey string

	// store is the underlying store for session data
	store sessions.Store
}

// NewStore initializes the session store.
func NewStore(name string, sessionKey string, secureCookie bool, underlyingStore sessions.Store) (*Store, error) {
	if underlyingStore == nil {
		redisStore, err := redistore.NewRediStore(10, "tcp", sessionStoreRedis, "", []byte(sessionCookieKeyRedis))
		if err != nil {
			return nil, err
		}
		redisStore.Options.Path = "/"
		redisStore.Options.HttpOnly = true
		redisStore.Options.Secure = secureCookie
		underlyingStore = redisStore
	}
	return &Store{
		name:       name,
		sessionKey: sessionKey,
		store:      underlyingStore,
	}, nil
}

// StartNewSession starts a new session with authentication for the given uid.
func (s *Store) StartNewSession(w http.ResponseWriter, r *http.Request, val []byte) error {
	s.DeleteSession(w, r)

	session, err := s.store.New(&http.Request{}, s.name) // workaround: not passing the request forces a new session
	if err != nil {
		log15.Error("error creating session", "error", err)
	}
	session.Values[s.sessionKey] = val
	if err = session.Save(r, w); err != nil {
		log15.Error("error saving session", "error", err)
	}

	return nil
}

// GetSession gets the current session token value, nil if it doesn't exist
func (s *Store) GetSession(r *http.Request) ([]byte, error) {
	session, err := s.store.Get(r, s.name)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	val := session.Values[s.sessionKey]
	if val == nil {
		return nil, nil
	} else if val, ok := val.([]byte); ok {
		return val, nil
	}
	return nil, fmt.Errorf("invalid value for session: %v", val)
}

// DeleteSession deletes the current session.
func (s *Store) DeleteSession(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.Get(r, s.name)
	if err != nil {
		log15.Error("error getting session", "error", err)
	}
	session.Options.MaxAge = -1
	if err = session.Save(r, w); err != nil {
		log15.Error("error saving session", "error", err)
	}
}

// Cookie returns the session cookie from the header of the given request.
func (s *Store) Cookie(r *http.Request) string {
	c, err := r.Cookie(s.name)
	if err != nil {
		return ""
	}
	return c.Value
}
