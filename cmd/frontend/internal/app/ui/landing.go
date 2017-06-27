package ui

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	htmpl "html/template"

	approuter "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/toprepos"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

var goSymbolReg = regexp.MustCompile("/info/GoPackage/(.+)$")

func shouldShadow(page string) bool {
	e := os.Getenv(page + "_LANDING_SHADOW_PERCENT")
	if e == "" {
		return false
	}
	p, err := strconv.Atoi(e)
	if err != nil {
		log15.Crit("landing: shouldShadow parsing "+page+"_LANDING_SHADOW_PERCENT", "error", err)
		return false
	}
	return rand.Uint32()%100 < uint32(p)
}

func shouldUseXlang(page string) bool {
	e := os.Getenv(page + "_LANDING_XLANG_PERCENT")
	if e == "" {
		return false
	}
	p, err := strconv.Atoi(e)
	if err != nil {
		log15.Crit("landing: shouldUseXlang parsing "+page+"_LANDING_XLANG_PERCENT", "error", err)
		return false
	}
	return rand.Uint32()%100 < uint32(p)
}

func init() { rand.Seed(time.Now().UnixNano()) }

// serveRepoLanding simply redirects the old (sourcegraph.com/<repo>/-/info) repo landing page
// URLs directly to the repo itself (sourcegraph.com/<repo>).
func serveRepoLanding(w http.ResponseWriter, r *http.Request) error {
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

func serveRepoIndex(w http.ResponseWriter, r *http.Request) error {
	lang := mux.Vars(r)["Lang"]
	var langDispName string
	var repos []toprepos.Repo
	switch strings.ToLower(lang) {
	case "go":
		repos = toprepos.GoRepos
		langDispName = "Go"
	case "java":
		repos = toprepos.JavaRepos
		langDispName = "Java"
	case "":
	default:
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: fmt.Errorf("language %q is not supported", lang)}
	}

	m := &meta{
		Title:       "Repositories",
		ShortTitle:  "Indexed repositories",
		Description: fmt.Sprintf("%s repositories indexed by Sourcegraph", langDispName),
		SEO:         true,
		Index:       true,
		Follow:      true,
	}

	return tmpl.Exec(r, w, "repoindex.html", http.StatusOK, nil, &struct {
		tmpl.Common
		Meta meta

		Lang         string
		LangDispName string
		Langs        []string
		Repos        []toprepos.Repo
	}{
		Meta: *m,

		Lang:         lang,
		LangDispName: langDispName,
		Langs:        []string{"Go"},
		Repos:        repos,
	})
}

type snippet struct {
	Code      htmpl.HTML
	SourceURL string
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
	http.Redirect(w, r, "/go/"+match[1], http.StatusTemporaryRedirect)
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
