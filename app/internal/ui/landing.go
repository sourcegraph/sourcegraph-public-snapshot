package ui

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"

	htmpl "html/template"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui/toprepos"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

// curlRepro returns curl reproduction instructions for an xlang request.
func curlRepro(mode, rootPath, method string, params interface{}) string {
	init := jsonrpc2.Request{
		ID:     jsonrpc2.ID{Num: 0},
		Method: "initialize",
	}
	init.SetParams(xlang.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{
			RootPath: rootPath,
		},
		Mode: mode,
	})
	req := jsonrpc2.Request{ID: jsonrpc2.ID{Num: 1}, Method: method}
	req.SetParams(params)
	shutdown := jsonrpc2.Request{ID: jsonrpc2.ID{Num: 2}, Method: "shutdown"}
	exit := jsonrpc2.Request{Method: "exit"}
	data, err := json.Marshal([]interface{}{init, req, shutdown, exit})
	if err != nil {
		log15.Crit("landing: curlRepro:", "error", err)
	}
	return fmt.Sprintf(`Reproduce with: curl --data '%s' https://sourcegraph.com/.api/xlang/%s -i`, data, method)
}

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

type defLandingData struct {
	tmpl.Common
	Meta             meta
	Description      *htmlutil.HTML
	RefSnippets      []*snippet
	ViewDefURL       string
	DefName          string // e.g. "func NewRouter"
	ShortDefName     string // e.g. "NewRouter"
	DefFileURL       string
	DefFileName      string
	DefEventProps    *defEventProps
	RefLocs          *sourcegraph.RefLocations
	TruncatedRefLocs bool
}

type defEventProps struct {
	DefLanguage  string `json:"def_language"`
	DefScheme    string `json:"def_scheme"`
	DefSource    string `json:"def_source"`
	DefContainer string `json:"def_container"`
	DefVersion   string `json:"def_version"`
	DefFile      string `json:"def_file"`
	DefName      string `json:"def_name"`
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

	repo, _, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		if errcode.IsHTTPErrorCode(err, http.StatusNotFound) {
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: err}
		}
		return errors.Wrap(err, "GetRepoAndRev")
	}

	// We only serve using srclib for the time being.
	legacyDefLandingCounter.Inc()
	data, err := queryLegacyDefLandingData(r, repo)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: err}
	}
	return tmpl.Exec(r, w, "deflanding.html", http.StatusOK, nil, data)
}

func legacyGenerateSymbolEventProps(def *sourcegraph.Def) *defEventProps {
	return &defEventProps{
		DefLanguage:  "go", // srclib def landing pages only ever had GoPackage unit types.
		DefScheme:    "git",
		DefSource:    def.Repo,
		DefVersion:   "",
		DefFile:      def.File,
		DefContainer: def.Unit,
		DefName:      def.Path,
	}
}

func queryLegacyDefLandingData(r *http.Request, repo *sourcegraph.Repo) (res *defLandingData, err error) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "queryLegacyDefLandingData")
	r = r.WithContext(ctx)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	vars := mux.Vars(r)
	// We don't support xlang yet on new commits for golang/go https://github.com/sourcegraph/sourcegraph/issues/2370
	// DefLanding is pretty only ever used on the default branch, so used old version
	if vars["Repo"] == "github.com/golang/go" {
		vars["Rev"] = "@838eaa738f2bc07c3706b96f9e702cb80877dfe1"
	}
	def, _, err := handlerutil.GetDefCommon(r.Context(), vars, &sourcegraph.DefGetOptions{Doc: true, ComputeLineRange: true})
	if err != nil {
		return nil, err
	}

	defSpec := sourcegraph.DefSpec{
		Repo:     repo.ID,
		CommitID: def.DefKey.CommitID,
		UnitType: def.DefKey.UnitType,
		Unit:     def.DefKey.Unit,
		Path:     def.DefKey.Path,
	}

	// get all caller repositories with counts (global refs)
	const (
		refLocRepoLimit = 3 // max 3 separate repos
		refLocFileLimit = 5 // max 5 files per repo
	)
	refLocs, err := backend.Defs.DeprecatedListRefLocations(r.Context(), &sourcegraph.DeprecatedDefsListRefLocationsOp{
		Def: defSpec,
		Opt: &sourcegraph.DeprecatedDefListRefLocationsOptions{
			// NOTE(mate): this has no effect at the moment
			ListOptions: sourcegraph.ListOptions{PerPage: refLocRepoLimit},
		},
	})
	if err != nil {
		return nil, err
	}
	// WORKAROUND(mate): because ListRefLocations ignores pagination options
	truncLen := len(refLocs.RepoRefs)
	if truncLen > refLocRepoLimit {
		truncLen = refLocRepoLimit
	}
	refLocs.RepoRefs = refLocs.RepoRefs[:truncLen]
	for _, repoRef := range refLocs.RepoRefs {
		if len(repoRef.Files) > refLocFileLimit {
			repoRef.Files = repoRef.Files[:refLocFileLimit]
		}
	}

	// fetch definition
	eventProps := legacyGenerateSymbolEventProps(def)
	viewDefURL := approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, int(def.StartLine))
	defFileURL := approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, 0)

	// fetch example
	refs, err := backend.Defs.DeprecatedListRefs(r.Context(), &sourcegraph.DeprecatedDefsListRefsOp{
		Def: defSpec,
		Opt: &sourcegraph.DeprecatedDefListRefsOptions{ListOptions: sourcegraph.ListOptions{PerPage: 3}},
	})
	if err != nil {
		return nil, err
	}
	var refSnippets []*snippet
	for _, ref := range refs.Refs {
		opt := &sourcegraph.RepoTreeGetOptions{
			ContentsAsString: true,
			GetFileOptions: sourcegraph.GetFileOptions{
				FileRange: sourcegraph.FileRange{
					StartByte: int64(ref.Start),
					EndByte:   int64(ref.End),
				},
				ExpandContextLines: 2,
			},
		}
		refRepo, err := backend.Repos.Resolve(r.Context(), &sourcegraph.RepoResolveOp{Path: ref.Repo})
		if err != nil {
			return nil, err
		}
		refEntrySpec := sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{Repo: refRepo.Repo, CommitID: ref.CommitID},
			Path:    ref.File,
		}
		refEntry, err := backend.RepoTree.Get(r.Context(), &sourcegraph.RepoTreeGetOp{Entry: refEntrySpec, Opt: opt})
		if err != nil {
			return nil, fmt.Errorf("could not get ref tree: %s", err)
		}
		refSnippets = append(refSnippets, &snippet{
			Code:      htmpl.HTML(refEntry.ContentsString),
			SourceURL: approuter.Rel.URLToBlob(ref.Repo, ref.CommitID, ref.File, int(refEntry.FileRange.StartLine+1)).String(),
		})
	}

	m := defMeta(def, trimRepo(repo.URI), false)
	m.SEO = true
	// Don't noindex pages with a canonical URL. See
	// https://www.seroundtable.com/archives/020151.html.
	m.CanonicalURL = canonicalRepoURL(
		conf.AppURL,
		getRouteName(r),
		mux.Vars(r),
		r.URL.Query(),
		repo.DefaultBranch,
		def.CommitID,
	)
	canonRev := isCanonicalRev(mux.Vars(r), repo.DefaultBranch)
	m.Index = allowRobots(repo) && shouldIndexDef(def) && canonRev

	return &defLandingData{
		Meta:             *m,
		Description:      def.DocHTML,
		RefSnippets:      refSnippets,
		DefEventProps:    eventProps,
		ViewDefURL:       viewDefURL.String(),
		DefName:          def.FmtStrings.DefKeyword + " " + def.FmtStrings.Name.ScopeQualified,
		ShortDefName:     def.Name,
		DefFileURL:       defFileURL.String(),
		DefFileName:      repo.URI + "/" + def.Def.File,
		RefLocs:          refLocs.Convert(),
		TruncatedRefLocs: refLocs.TotalRepos > int32(len(refLocs.RepoRefs)),
	}, nil
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
