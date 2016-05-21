// Package router contains the URL router for the Git "smart" protocol
// HTTP handlers.
package router

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

const (
	GitInfoRefs    = "git.info-refs"
	GitUploadPack  = "git.upload-pack"
	GitReceivePack = "git.receive-pack"
)

// New creates a new Git HTTP router.
func New(base *mux.Router) *mux.Router {
	if base == nil {
		base = mux.NewRouter()
	}

	m := base.MatcherFunc(func(req *http.Request, rt *mux.RouteMatch) bool {
		userAgent := req.Header.Get("User-Agent")
		return strings.HasPrefix(strings.ToLower(userAgent), "git/")
	}).Subrouter()
	m.Path("/" + routevar.Repo + "/info/refs").Methods("GET").Name(GitInfoRefs)
	m.Path("/" + routevar.Repo + "/git-upload-pack").Methods("POST").Name(GitUploadPack)
	m.Path("/" + routevar.Repo + "/git-receive-pack").Methods("POST").Name(GitReceivePack)

	return base
}
