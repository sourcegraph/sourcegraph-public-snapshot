package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/russross/blackfriday"
	"github.com/sourcegraph/sourcegraph-go/pkg/lsp"

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
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

type defDescr struct {
	Def       *sourcegraph.Def
	RefCount  int32
	LandURL   string
	SourceURL string
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

func serveRepoLanding(w http.ResponseWriter, r *http.Request) error {
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
		return err
	}

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

	var defDescrs []defDescr
	for _, defResult := range results.DefResults {
		def := &defResult.Def
		handlerutil.ComputeDocHTML(def)
		defDescrs = append(defDescrs, defDescr{
			Def:       def,
			RefCount:  defResult.RefCount,
			LandURL:   approuter.Rel.URLToDefLanding(def.DefKey).String(),
			SourceURL: approuter.Rel.URLToDefKey(def.DefKey).String(),
		})
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
		Defs:            defDescrs,
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
	RefLocs          *sourcegraph.RefLocations
	TruncatedRefLocs bool
}

func serveDefLanding(w http.ResponseWriter, r *http.Request) error {
	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		if errcode.IsHTTPErrorCode(err, http.StatusNotFound) {
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: err}
		}
		return errors.Wrap(err, "GetRepoAndRev")
	}

	migrated, err := backend.Defs.HasMigrated(r.Context(), repo.URI)
	if err != nil {
		// Just log, so we fallback to legacy.
		log15.Crit("Defs.HasMigrated", "error", err)
	}

	var data *defLandingData
	if migrated {
		data, err = queryDefLandingData(r, repo, repoRev)
		if err != nil {
			return err
		}
	} else {
		// Fallback to legacy / srclib data.
		data, err = queryLegacyDefLandingData(r, repo)
		if err != nil {
			return err
		}
	}
	return tmpl.Exec(r, w, "deflanding.html", http.StatusOK, nil, data)
}

func queryDefLandingData(r *http.Request, repo *sourcegraph.Repo, repoRev sourcegraph.RepoRevSpec) (*defLandingData, error) {
	defSpec := routevar.ToDefAtRev(mux.Vars(r))
	language := "go" // TODO(slimsag): long term, add to route

	// Lookup the definition based on the legacy srclib defkey in the page URL.
	rootPath := "git://" + defSpec.Repo + "?" + repoRev.CommitID
	var symbols []lsp.SymbolInformation
	err := xlang.OneShotClientRequest(r.Context(), language, rootPath, "workspace/symbol", lsp.WorkspaceSymbolParams{
		// TODO(slimsag): before merge, performance for golang/go here is not
		// good. Allow specifying file URIs as a query filter. Sucks a bit that
		// textDocument/definition won't give us the Name/ContainerName that we
		// need!
		Query: "", // all symbols
	}, &symbols)

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
		var (
			defContainerName string
			split            = strings.Split(defSpec.Path, "/")
			defName          = split[0]
		)
		if len(split) == 2 {
			defContainerName, defName = split[0], split[1]
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
		return nil, fmt.Errorf("could not finding matching symbol info")
	}

	// Create links to the definition.
	symURI, err := url.Parse(symbol.Location.URI)
	if err != nil {
		return nil, errors.Wrap(err, "parsing symbol location URI")
	}
	defFileURL := "/" + symURI.Host + symURI.Path + "/-/blob/" + path.Clean(symURI.Fragment)
	defFileName := symURI.Host + path.Join(symURI.Path, symURI.Fragment)
	viewDefURL := fmt.Sprintf("%s#L%d", defFileURL, symbol.Location.Range.Start.Line+1)

	// Create metadata titles.
	shortTitle := strings.Join([]string{symbol.ContainerName, symbol.Name}, ".")
	title := repoPageTitle(trimRepo(symURI.Host+symURI.Path), shortTitle)

	// Determine the definition title.
	var hover lsp.Hover
	err = xlang.OneShotClientRequest(r.Context(), language, rootPath, "textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: symbol.Location.URI},
		Position: lsp.Position{
			Line:      symbol.Location.Range.Start.Line,
			Character: symbol.Location.Range.Start.Character,
		},
	}, &hover)
	hoverTitle := hover.Contents[0].Value

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
	shouldIndexSymbol := len(symbol.Name) >= 3 && (symbol.Kind == lsp.SKClass || symbol.Kind == lsp.SKConstructor || symbol.Kind == lsp.SKFunction || symbol.Kind == lsp.SKInterface || symbol.Kind == lsp.SKMethod)

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
							// Note: we can expand the lines without bounds
							// checking because RepoTree.Get is smart.
							StartLine: int64(ref.Positions[0].Start.Line+1) - 2,
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
			Index:        allowRobots(repo) && shouldIndexSymbol && canonRev,
		},
		// TODO(slimsag): before merge, get descriptions from LSP hover
		//Description: htmlutil.Sanitize("description"),
		RefSnippets:      refSnippets,
		ViewDefURL:       viewDefURL,
		DefName:          hoverTitle,
		ShortDefName:     symbol.Name,
		DefFileURL:       defFileURL,
		DefFileName:      defFileName,
		RefLocs:          refLocs,
		TruncatedRefLocs: refLocs.TotalSources > len(refLocs.SourceRefs),
	}, nil
}

func queryLegacyDefLandingData(r *http.Request, repo *sourcegraph.Repo) (*defLandingData, error) {
	vars := mux.Vars(r)
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
	viewDefURL := approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, int(def.StartLine)).String()
	defFileURL := approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, 0).String()

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
		ViewDefURL:       viewDefURL,
		DefName:          def.FmtStrings.DefKeyword + " " + def.FmtStrings.Name.ScopeQualified,
		ShortDefName:     def.Name,
		DefFileURL:       defFileURL,
		DefFileName:      repo.URI + "/" + def.Def.File,
		RefLocs:          refLocs.Convert(),
		TruncatedRefLocs: refLocs.TotalRepos > int32(len(refLocs.RepoRefs)),
	}, nil
}
