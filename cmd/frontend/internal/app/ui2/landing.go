package ui2

import (
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	approuter "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

var goSymbolReg = regexp.MustCompile("/info/GoPackage/(.+)$")

// serveRepoLanding simply redirects the old (sourcegraph.com/<repo>/-/info) repo landing page
// URLs directly to the repo itself (sourcegraph.com/<repo>).
func serveRepoLanding(w http.ResponseWriter, r *http.Request) error {
	legacyRepoLandingCounter.Inc()

	repo, rev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		if errcode.IsHTTPErrorCode(err, http.StatusNotFound) {
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: err}
		}
		return errors.Wrap(err, "GetRepoAndRev")
	}
	http.Redirect(w, r, approuter.Rel.URLToRepoRev(repo.URI, rev.CommitID).String(), http.StatusMovedPermanently)
	return nil
}

func serveDefLanding(w http.ResponseWriter, r *http.Request) (err error) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "serveDefLanding")
	r = r.WithContext(ctx)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	legacyDefLandingCounter.Inc()

	match := goSymbolReg.FindStringSubmatch(r.URL.Path)
	if match == nil {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: err}
	}
	http.Redirect(w, r, "/go/"+match[1], http.StatusMovedPermanently)
	return nil
}

var legacyDefLandingCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "legacy_def_landing",
	Help:      "Number of times a legacy def landing page has been served.",
})

var legacyRepoLandingCounter = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Name:      "legacy_repo_landing",
	Help:      "Number of times a legacy repo landing page has been served.",
})

func init() {
	prometheus.MustRegister(legacyDefLandingCounter)
	prometheus.MustRegister(legacyRepoLandingCounter)
}
