package ui

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/jaytaylor/html2text"
	"github.com/microcosm-cc/bluemonday"
	"github.com/neelance/parallel"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/russross/blackfriday"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"

	htmpl "html/template"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui/toprepos"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
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

type defDescr struct {
	Title    string
	DocText  string
	RefCount int
	URL      string

	// These two fields are unused today, but specify the number of files
	// (across all sources/repos) and the number of sources (repos) that
	// reference the definition.
	FileCount, SourcesCount int

	// sortIndex is a private field use for sorting the concurrently retrieved
	// descriptions.
	sortIndex int
}

type sortedDefDescr []defDescr

func (s sortedDefDescr) Len() int           { return len(s) }
func (s sortedDefDescr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortedDefDescr) Less(i, j int) bool { return s[i].sortIndex < s[j].sortIndex }

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

func serveRepoLanding(w http.ResponseWriter, r *http.Request) (err error) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "serveRepoLanding")
	r = r.WithContext(ctx)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	vars := mux.Vars(r)

	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), vars)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: err}
	}

	// terminate early on non-Go repos
	if repo.Language != "Go" {
		http.Error(w, "404 - Page not found. (No landing page for non-Go repo.)", http.StatusNotFound)
		return nil
	}

	repoURL := approuter.Rel.URLToRepo(repo.URI).String()

	var sanitizedREADME []byte
	readmeEntry, err := backend.RepoTree.Get(r.Context(), &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: repoRev,
			Path:    "README.md",
		},
	})
	if err != nil && errcode.Code(err) != legacyerr.NotFound {
		return err
	} else if err == nil {
		sanitizedREADME = bluemonday.UGCPolicy().SanitizeBytes(blackfriday.MarkdownCommon(readmeEntry.Contents))
	}

	var data []defDescr
	err = errors.New("xlang repo landing disabled")
	if shouldUseXlang("REPO") {
		data, err = queryRepoLandingData(r, repo)
		if err != nil {
			// Just log, so we fallback to legacy in the event of catastrophic failure.
			log15.Crit("queryRepoLandingData", "error", err, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
	} else if shouldShadow("REPO") {
		go func() {
			_, err := queryRepoLandingData(r, repo)
			if err != nil {
				log15.Crit("queryRepoLandingData: shadow", "error", err, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
				ext.Error.Set(span, true)
				span.SetTag("err", err.Error())
			}
		}()
	}

	// If the new system failed for some reason OR if we don't have 5 top
	// definitions, then fallback to the legacy srclib data.
	if err != nil || len(data) < 5 {
		// Fallback to legacy / srclib data.
		legacyRepoLandingCounter.Inc()
		legacyData, legacyErr := queryLegacyRepoLandingData(r, repo)
		if legacyErr != nil && err != nil {
			// Only return an error if both systems error'd out.
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: fmt.Errorf("multiple errors: queryRepoLandingData=%v queryLegacyRepoLandingData=%v", err, legacyErr)}
		}
		if legacyData != nil {
			data = legacyData
		}
	}

	m := repoMeta(repo)
	m.SEO = true

	return tmpl.Exec(r, w, "repolanding.html", http.StatusOK, nil, &struct {
		tmpl.Common
		Meta meta

		SanitizedREADME string
		Repo            *sourcegraph.Repo
		RepoRev         sourcegraph.RepoRevSpec
		RepoURL         string
		Defs            []defDescr
	}{
		Meta:            *m,
		SanitizedREADME: string(sanitizedREADME),
		Repo:            repo,
		RepoRev:         repoRev,
		RepoURL:         repoURL,
		Defs:            data,
	})
}

func queryRepoLandingData(r *http.Request, repo *sourcegraph.Repo) (res []defDescr, err error) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "queryRepoLandingData")
	r = r.WithContext(ctx)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	language := "go" // TODO(slimsag): long term, add to route

	// Query information about the top definitions in the repo.
	topDefs, err := backend.Defs.TopDefs(r.Context(), sourcegraph.TopDefsOptions{
		Source: repo.URI,
		Limit:  5,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Defs.TopDefs")
	}

	run := parallel.NewRun(32)
	var (
		descMu sync.RWMutex
		desc   []defDescr
	)
	for i, d := range topDefs.SourceDefs {
		d := d
		sortIndex := i
		run.Acquire()
		go func() {
			defer run.Release()

			// Lookup the definition based on the source / definition (Name, ContainerName)
			// pair.
			var (
				rootPath = "git://" + repo.URI + "?" + repo.DefaultBranch
				symbols  []lsp.SymbolInformation
				method   string
				params   interface{}
			)
			if path.Ext(d.DefFile) != "" {
				// d.DefFile is a file, so use textDocument/documentSymbol
				method = "textDocument/documentSymbol"
				params = lsp.DocumentSymbolParams{
					TextDocument: lsp.TextDocumentIdentifier{
						URI: rootPath + "#" + d.DefFile,
					},
				}
			} else {
				// d.DefFile is a directory, so use workspace/symbol with the
				// `dir:` filter.
				method = "workspace/symbol"
				params = lsp.WorkspaceSymbolParams{
					Query: fmt.Sprintf("dir:%s %s", d.DefFile, d.DefName),
					Limit: 100,
				}
			}
			err = xlang.UnsafeOneShotClientRequest(r.Context(), language, rootPath, method, params, &symbols)
			if err != nil {
				run.Error(errors.Wrap(err, "LSP "+method))
				return
			}

			// Find the matching symbol.
			var symbol *lsp.SymbolInformation
			for _, sym := range symbols {
				if sym.Name != d.DefName || sym.ContainerName != d.DefContainerName {
					continue
				}
				withoutFile, err := url.Parse(sym.Location.URI)
				if err != nil {
					run.Error(errors.Wrap(err, "parsing symbol location URI"))
					return
				}
				withoutFile.Fragment = ""
				if withoutFile.String() != rootPath {
					continue
				}
				symbol = &sym
				break
			}
			if symbol == nil {
				msg := "queryRepoLandingData: no symbol info matching top def from global references"
				log15.Warn(msg, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
				log15.Warn(curlRepro(language, rootPath, method, params))
				span.LogEvent(msg)
				span.SetTag("missing", "symbol")
				span.LogEvent(curlRepro(language, rootPath, method, params))
				return
			}

			// Determine the definition title.
			var hover lsp.Hover
			method = "textDocument/hover"
			params = lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: symbol.Location.URI},
				Position: lsp.Position{
					Line:      symbol.Location.Range.Start.Line,
					Character: symbol.Location.Range.Start.Character,
				},
			}
			err = xlang.UnsafeOneShotClientRequest(r.Context(), language, rootPath, method, params, &hover)
			if len(hover.Contents) == 0 {
				msg := "queryRepoLandingData: LSP textDocument/hover returned no contents"
				log15.Warn(msg, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
				log15.Warn(curlRepro(language, rootPath, method, params))
				span.LogEvent(msg)
				span.SetTag("missing", "hover")
				span.LogEvent(curlRepro(language, rootPath, method, params))
				return
			}

			hoverTitle := hover.Contents[0].Value
			var hoverDesc string
			for _, s := range hover.Contents {
				if s.Language == "markdown" {
					hoverDesc = s.Value
					break
				}
			}

			u, err := approuter.Rel.URLToLegacyDefLanding(*symbol)
			if err != nil {
				run.Error(err)
				return
			}
			descMu.Lock()
			defer descMu.Unlock()
			desc = append(desc, defDescr{
				Title:        hoverTitle,
				DocText:      hoverDesc,
				RefCount:     d.Refs,
				FileCount:    d.Files,
				SourcesCount: d.Sources,
				URL:          u,
				sortIndex:    sortIndex,
			})
		}()
	}
	if err := run.Wait(); err != nil {
		return nil, err
	}
	sort.Sort(sortedDefDescr(desc))
	return desc, nil
}

func defDocText(def *sourcegraph.Def) string {
	for _, doc := range def.Docs {
		if doc.Format == "text/plain" {
			return doc.Data
		}
	}
	if plain, err := html2text.FromString(def.DocHTML.HTML); err == nil {
		return plain
	}
	for _, doc := range def.Docs {
		if doc.Format == "markdown" {
			if plain, err := html2text.FromString(doc.Data); err == nil {
				return plain
			}
		}
	}
	return ""
}

func queryLegacyRepoLandingData(r *http.Request, repo *sourcegraph.Repo) (res []defDescr, err error) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "queryLegacyRepoLandingData")
	r = r.WithContext(ctx)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	results, err := backend.Search.Search(r.Context(), &sourcegraph.SearchOp{
		Opt: &sourcegraph.SearchOptions{
			Repos:        []int32{repo.ID},
			Languages:    []string{"Go"},
			NotKinds:     []string{"package"},
			IncludeRepos: false,
			ListOptions:  sourcegraph.ListOptions{PerPage: 20},
		},
	})
	if err != nil {
		return nil, err
	}

	var defDescrs []defDescr
	for _, defResult := range results.DefResults {
		def := &defResult.Def
		handlerutil.ComputeDocHTML(def)
		f := def.FmtStrings
		defDescrs = append(defDescrs, defDescr{
			Title:    fmt.Sprintf("%s %s %s%s", f.DefKeyword, f.Name.ScopeQualified, f.NameAndTypeSeparator, f.Type.Unqualified),
			DocText:  defDocText(def),
			RefCount: int(defResult.RefCount),
			URL:      approuter.Rel.URLToDefLanding(def.DefKey).String(),
		})
	}
	return defDescrs, nil
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

	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		if errcode.IsHTTPErrorCode(err, http.StatusNotFound) {
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: err}
		}
		return errors.Wrap(err, "GetRepoAndRev")
	}

	// TODO(slimsag): see https://github.com/sourcegraph/sourcegraph/issues/2021
	var data *defLandingData
	err = errors.New("xlang def landing disabled")
	if shouldUseXlang("DEF") {
		data, err = queryDefLandingData(r, repo, repoRev)
		if err != nil {
			// Just log, so we fallback to legacy in the event of catastrophic failure.
			log15.Crit("queryDefLandingData", "error", err, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
	} else if shouldShadow("DEF") {
		go func() {
			_, err := queryDefLandingData(r, repo, repoRev)
			if err != nil {
				log15.Crit("queryDefLandingData: shadow", "error", err, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
				ext.Error.Set(span, true)
				span.SetTag("err", err.Error())
			}
		}()
	}

	// If the new system failed for some reason OR if we don't have 3 sources
	// referencing this page's definition, then fallback to the legacy srclib
	// data.
	if err != nil || len(data.RefLocs.SourceRefs) < 3 {
		legacyDefLandingCounter.Inc()
		legacyData, legacyErr := queryLegacyDefLandingData(r, repo)
		if legacyErr != nil && err != nil {
			// Only return an error if both systems error'd out.
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: fmt.Errorf("multiple errors: queryDefLandingData=%v queryLegacyDefLandingData=%v", err, legacyErr)}
		}
		if legacyData != nil {
			data = legacyData
		}
	}
	return tmpl.Exec(r, w, "deflanding.html", http.StatusOK, nil, data)
}

// generateSymbolEventProps generates symbol event logging properties
// It panics if symbol.Location.URI is unparsable.
func generateSymbolEventProps(language string, symbol *lsp.SymbolInformation) *defEventProps {
	symURI, err := url.Parse(symbol.Location.URI)
	if err != nil {
		panic(err)
	}
	return &defEventProps{
		DefLanguage:  language,
		DefScheme:    symURI.Scheme,
		DefSource:    (symURI.Host + symURI.Path),
		DefVersion:   symURI.RawQuery,
		DefFile:      symURI.Fragment,
		DefContainer: symbol.ContainerName,
		DefName:      symbol.Name,
	}
}

func queryDefLandingData(r *http.Request, repo *sourcegraph.Repo, repoRev sourcegraph.RepoRevSpec) (res *defLandingData, err error) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "queryDefLandingData")
	r = r.WithContext(ctx)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	defSpec := routevar.ToDefAtRev(mux.Vars(r))
	language := "go" // TODO(slimsag): long term, add to route

	var (
		defContainerName string
		split            = strings.Split(defSpec.Path, "/")
		defName          = split[0]
	)
	if len(split) == 2 {
		defContainerName, defName = split[0], split[1]
	} else {
		split := strings.Split(defSpec.Unit, "/")
		defContainerName = split[len(split)-1]
	}

	// Lookup the definition based on the legacy srclib defkey in the page URL.
	rootPath := "git://" + defSpec.Repo + "?" + repoRev.CommitID
	var symbols []lsp.SymbolInformation
	method := "workspace/symbol"
	params := lsp.WorkspaceSymbolParams{
		Query: defName,
		Limit: 100,
	}
	err = xlang.UnsafeOneShotClientRequest(r.Context(), language, rootPath, method, params, &symbols)

	// Find the matching symbol.
	var symbol *lsp.SymbolInformation
	for _, sym := range symbols {
		withoutFile, err := url.Parse(sym.Location.URI)
		if err != nil {
			return nil, errors.Wrap(err, "parsing symbol location URI")
		}
		withoutFile.Fragment = ""
		if withoutFile.String() != "git://"+defSpec.Repo+"?"+repoRev.CommitID {
			continue
		}
		// Note: We must check defContainerName is empty because xlang sets the
		// container name to the package name if there is none, in contrast
		// srclib would leave it empty. This is a sort of fuzzy matching, but
		// always works.
		containersEqual := defContainerName == "" || (sym.ContainerName == defContainerName)
		if !containersEqual || sym.Name != defName {
			continue
		}
		symbol = &sym
		break
	}
	if symbol == nil {
		msg := "queryDefLandingData: could not find matching symbol info"
		log15.Warn(msg, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
		log15.Warn(curlRepro(language, rootPath, method, params))
		span.LogEvent(msg)
		span.SetTag("missing", "symbol")
		span.LogEvent(curlRepro(language, rootPath, method, params))
		return nil, errors.New("LSP workspace/symbol did not return matching symbol info")
	}

	// Create links to the definition.
	symURI, err := url.Parse(symbol.Location.URI)
	if err != nil {
		return nil, errors.Wrap(err, "parsing symbol location URI")
	}
	defFileName := symURI.Host + path.Join(symURI.Path, symURI.Fragment)
	repoURI := symURI.Host + symURI.Path

	eventProps := generateSymbolEventProps(language, symbol)
	defFileURL := approuter.Rel.URLToBlob(repoURI, "", path.Clean(symURI.Fragment), 0)
	viewDefURL := approuter.Rel.URLToBlob(repoURI, "", path.Clean(symURI.Fragment), symbol.Location.Range.Start.Line+1)

	// Create metadata titles.
	shortTitle := strings.Join([]string{symbol.ContainerName, symbol.Name}, ".")
	title := repoPageTitle(trimRepo(symURI.Host+symURI.Path), shortTitle)

	// Determine the definition title.
	var hover lsp.Hover
	method = "textDocument/hover"
	hoverParams := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: symbol.Location.URI},
		Position: lsp.Position{
			Line:      symbol.Location.Range.Start.Line,
			Character: symbol.Location.Range.Start.Character,
		},
	}

	err = xlang.UnsafeOneShotClientRequest(r.Context(), language, rootPath, method, hoverParams, &hover)
	if len(hover.Contents) == 0 {
		msg := "queryDefLandingData: LSP textDocument/hover returned no contents"
		log15.Crit(msg, "trace", traceutil.SpanURL(opentracing.SpanFromContext(r.Context())))
		log15.Crit(curlRepro(language, rootPath, method, hoverParams))
		span.LogEvent(msg)
		span.SetTag("missing", "hover")
		span.LogEvent(curlRepro(language, rootPath, method, hoverParams))
		return nil, errors.New("LSP textDocument/hover returned no contents")
	}

	hoverTitle := hover.Contents[0].Value
	var hoverDesc string
	for _, s := range hover.Contents {
		if s.Language == "markdown" {
			hoverDesc = string(blackfriday.MarkdownCommon([]byte(s.Value)))
			break
		}
	}

	// Determine canonical URL and whether the symbol shold be indexed.
	canonicalURL := canonicalRepoURL(
		conf.AppURL,
		getRouteName(r),
		mux.Vars(r),
		r.URL.Query(),
		repo.DefaultBranch,
		repoRev.CommitID,
	)
	canonRev := isCanonicalRev(mux.Vars(r), repo.DefaultBranch)

	goodName := len(symbol.Name) >= 3
	goodKind := symbol.Kind == lsp.SKClass || symbol.Kind == lsp.SKConstructor || symbol.Kind == lsp.SKFunction || symbol.Kind == lsp.SKInterface || symbol.Kind == lsp.SKMethod
	goodDocs := len(hoverDesc) >= 20
	goodSymbol := goodName && goodKind && goodDocs

	// Request up to 5 files for up to 3 sources (e.g. repos) that reference
	// the definition.
	refLocs, err := backend.Defs.RefLocations(r.Context(), sourcegraph.RefLocationsOptions{
		Sources:       3,
		Files:         5,
		Source:        symURI.Host + symURI.Path,
		Name:          symbol.Name,
		ContainerName: symbol.ContainerName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Defs.RefLocations")
	}

	var (
		maxUsageExamples = 3
		exampleIndex     = 0
		refSnippets      []*snippet
	)
	for _, sourceRef := range refLocs.SourceRefs {
		if exampleIndex > maxUsageExamples {
			break
		}
		for _, ref := range sourceRef.FileRefs {
			exampleIndex++
			if exampleIndex > maxUsageExamples {
				break
			}
			refRepo, err := backend.Repos.Resolve(r.Context(), &sourcegraph.RepoResolveOp{Path: ref.Source})
			if err != nil {
				return nil, errors.Wrap(err, "Repos.Resolve")
			}
			startLine := int64(ref.Positions[0].Start.Line+1) - 2
			if startLine < 0 {
				startLine = 0
			}
			refEntry, err := backend.RepoTree.Get(r.Context(), &sourcegraph.RepoTreeGetOp{
				Entry: sourcegraph.TreeEntrySpec{
					// TODO: does ref.Version need resolving?
					RepoRev: sourcegraph.RepoRevSpec{Repo: refRepo.Repo, CommitID: ref.Version},
					Path:    ref.File,
				},
				Opt: &sourcegraph.RepoTreeGetOptions{
					ContentsAsString: true,
					GetFileOptions: sourcegraph.GetFileOptions{
						FileRange: sourcegraph.FileRange{
							// Note: we can expand the end line without bounds
							// checking because RepoTree.Get is smart.
							StartLine: startLine,
							EndLine:   int64(ref.Positions[0].End.Line+1) + 2,
						},
						ExpandContextLines: 2,
					},
					NoSrclibAnns: true,
				},
			})
			if err != nil {
				return nil, errors.Wrap(err, "RepoTree.Get")
			}
			refSnippets = append(refSnippets, &snippet{
				Code:      htmpl.HTML(refEntry.ContentsString),
				SourceURL: approuter.Rel.URLToBlob(ref.Source, ref.Version, ref.File, ref.Positions[0].Start.Line+1).String(),
			})
		}
	}

	return &defLandingData{
		Meta: meta{
			SEO:        true,
			Title:      title,
			ShortTitle: shortTitle,

			// TODO(slimsag): Gather additional description information from
			// hover once docs are available.
			Description: "Go usage examples and docs for " + hoverTitle,

			// Don't noindex pages with a canonical URL. See
			// https://www.seroundtable.com/archives/020151.html.
			CanonicalURL: canonicalURL,
			Index:        allowRobots(repo) && goodSymbol && canonRev,
		},
		Description:      htmlutil.Sanitize(hoverDesc),
		RefSnippets:      refSnippets,
		DefEventProps:    eventProps,
		ViewDefURL:       viewDefURL.String(),
		DefName:          hoverTitle,
		ShortDefName:     symbol.Name,
		DefFileURL:       defFileURL.String(),
		DefFileName:      defFileName,
		RefLocs:          refLocs,
		TruncatedRefLocs: refLocs.TotalSources > len(refLocs.SourceRefs),
	}, nil
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
			NoSrclibAnns: true,
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
