package gologin

import (
	"net/http"
)

// DefaultFailureHandler responds with a 400 status code and message parsed
// from the ctx.
var DefaultFailureHandler = http.HandlerFunc(failureHandler)

func failureHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	err := ErrorFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// should be unreachable, ErrorFromContext always returns some non-nil error
	http.Error(w, "", http.StatusBadRequest)
}
