package publicrestapi

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func NewHandler(
	apiHandler http.Handler,
) http.Handler {
	logger := sglog.Scoped("publicrestapi")

	m := mux.NewRouter().PathPrefix("/api/").Subrouter()
	m.StrictSlash(true)
	m.Use(trace.Route)

	jsonHandler := httpapi.JsonMiddleware(&httpapi.ErrorHandler{
		Logger: logger,
		// Only display error message to admins when in debug mode, since it
		// may contain sensitive info (like API keys in net/http error
		// messages).
		WriteErrBody: env.InsecureDev,
	})

	registerOpenAIRoutes(m, logger, apiHandler, jsonHandler)

	return m
}

func registerOpenAIRoutes(m *mux.Router, logger sglog.Logger, apiHandler http.Handler, jsonHandler func(func(http.ResponseWriter, *http.Request) error) http.Handler) {
	router := m.PathPrefix("/openai/v1/").Subrouter()
	// Converts `Authorization: Bearer <token>` to `Authorization: token <token>` to match Sourcegraph APIs.
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				r.Header.Set("Authorization", "token "+strings.TrimPrefix(auth, "Bearer "))
			}
			next.ServeHTTP(w, r)
		})
	})

	router.Path("/chat/completions").Methods("POST").Handler(jsonHandler(serveOpenAIChatCompletions(logger, apiHandler)))
}
