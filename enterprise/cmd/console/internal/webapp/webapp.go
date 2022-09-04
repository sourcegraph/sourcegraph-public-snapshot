package webapp

import (
	"fmt"

	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/sessions"
	"github.com/sourcegraph/log"
)

type Config struct {
	ExternalURL url.URL // root external URL of the web app

	SessionKey string // secret key for authenticating session data

	Logger *log.Logger // logger for webapp
}

// New returns an HTTP handler that serves the web app as configured.
func New(config Config) *Handler {
	handler := &Handler{
		config: config,
	}

	// Sessions
	if config.SessionKey == "" {
		panic("invalid empty session key")
	}
	sessionStore := sessions.NewCookieStore([]byte(config.SessionKey))
	// Reuse handler's cookie options for session cookies.
	cookieOptions := handler.authCookie("", "", time.Time{})
	sessionStore.Options = &sessions.Options{
		Path:     cookieOptions.Path,
		Secure:   cookieOptions.Secure,
		HttpOnly: cookieOptions.HttpOnly,
		SameSite: cookieOptions.SameSite,
	}
	handler.session = sessionStore

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.serveRoot)
	mux.HandleFunc("/instances", handler.serveInstances)

	handler.mux = mux
	return handler
}

// Handler implements http.Handler.
type Handler struct {
	Logger log.Logger

	config Config

	session sessions.Store

	mux http.Handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Protect against CSRF.

	// TODO(sqs): can we use an existing http handler wrapper?
	if requestOrigin := r.Header.Get("Origin"); requestOrigin != "" && !sameOrigin(requestOrigin, h.config.ExternalURL) {
		h.handlerError(w, "bad request origin", fmt.Errorf("got origin %q, want %q", requestOrigin, h.config.ExternalURL.String()))
		return
	}

	h.mux.ServeHTTP(w, r)
}

func sameOrigin(requestOrigin string, wantOrigin url.URL) bool {
	u, err := url.Parse(requestOrigin)
	if err != nil {
		return false
	}
	return u.Scheme == wantOrigin.Scheme && u.Host == wantOrigin.Host
}

func (h *Handler) handlerError(w http.ResponseWriter, operation string, err error) {
	http.Error(w, "error "+operation, http.StatusInternalServerError)
	h.Logger.Warn("HTTP error", log.String("op", operation), log.Error(err))
}
