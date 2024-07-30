package publicrestapi

import (
	"net/http"

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

	m.Path("/openai/v1/chat/completions").Methods("POST").Handler(jsonHandler(serveOpenAIChatCompletions(logger, apiHandler)))

	return m
}
