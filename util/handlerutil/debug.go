package handlerutil

import (
	"net/http"
	"os"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

// DebugMode returns whether debugging information should be emitted
// with the request. It assumes that ActorMiddleware has already run.
func DebugMode(r *http.Request) bool {
	if v, _ := strconv.ParseBool(os.Getenv("DEBUG")); v {
		return true
	}
	if auth.ActorFromContext(httpctx.FromRequest(r)).Admin {
		return true
	}
	return false
}
