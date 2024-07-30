package publicrestapi

import (
	"net/http"

	sglog "github.com/sourcegraph/log"
)

// serveOpenAIChatCompletions is a handler for the OpenAI /v1/chat/completions endpoint.
func serveOpenAIChatCompletions(logger sglog.Logger, apiHandler http.Handler) func(w http.ResponseWriter, r *http.Request) (err error) {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		panic("unimplemented")
	}
}
