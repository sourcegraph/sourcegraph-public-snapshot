package app

import (
	"net/http"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func serveGDDORefs(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	u := router.Rel.URLToDefSubroute(router.DefExamples, graph.DefKey{
		Repo:     q.Get("repo"),
		UnitType: "GoPackage",
		Unit:     q.Get("pkg"),
		Path:     strings.Replace(q.Get("def"), ".", "/", -1),
	})
	http.Redirect(w, r, u.String(), http.StatusMovedPermanently)
	return nil
}
