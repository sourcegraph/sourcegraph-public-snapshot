package ui

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"

	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/snippet"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui/toprepos"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
	"sourcegraph.com/sqs/pbtypes"
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
	cl := handlerutil.Client(r)
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

	results, err := cl.Search.Search(r.Context(), &sourcegraph.SearchOp{
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
	readmeEntry, err := cl.RepoTree.Get(r.Context(), &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: repoRev,
			Path:    "README.md",
		},
	})
	if err != nil && errcode.GRPC(err) != codes.NotFound {
		return err
	} else if err == nil {
		sanitizedREADME = bluemonday.UGCPolicy().SanitizeBytes(blackfriday.MarkdownCommon(readmeEntry.Contents))
	}

	var defDescrs []defDescr
	for _, defResult := range results.DefResults {
		def := &defResult.Def
		htmlutil.ComputeDocHTML(def)
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

func serveDefLanding(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)
	vars := mux.Vars(r)

	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), vars)
	if err != nil {
		return err
	}

	var def *sourcegraph.Def
	var refLocs *sourcegraph.RefLocationsList
	var defEntry *sourcegraph.TreeEntry
	var defSnippet *snippet.Snippet
	var refSnippets []*snippet.Snippet
	var viewDefURL, defFileURL string

	if def == nil {
		def, _, err = handlerutil.GetDefCommon(r.Context(), vars, &sourcegraph.DefGetOptions{Doc: true, ComputeLineRange: true})
		if err != nil {
			return err
		}

		defSpec := sourcegraph.DefSpec{
			Repo:     repo.ID,
			CommitID: def.DefKey.CommitID,
			UnitType: def.DefKey.UnitType,
			Unit:     def.DefKey.Unit,
			Path:     def.DefKey.Path,
		}

		// get all caller repositories with counts (global refs)
		const reflocRepoLimit = 5
		refLocs, err = cl.Defs.ListRefLocations(r.Context(), &sourcegraph.DefsListRefLocationsOp{
			Def: defSpec,
			Opt: &sourcegraph.DefListRefLocationsOptions{
				// NOTE(mate): this has no effect at the moment
				ListOptions: sourcegraph.ListOptions{PerPage: reflocRepoLimit},
			},
		})
		if err != nil {
			return err
		}
		// WORKAROUND(mate): because ListRefLocations ignores pagination options
		truncLen := len(refLocs.RepoRefs)
		if truncLen > reflocRepoLimit {
			truncLen = reflocRepoLimit
		}
		refLocs.RepoRefs = refLocs.RepoRefs[:truncLen]

		// fetch definition
		entrySpec := sourcegraph.TreeEntrySpec{
			RepoRev: repoRev,
			Path:    def.Def.File,
		}
		opt := sourcegraph.RepoTreeGetOptions{
			ContentsAsString: true,
			GetFileOptions: sourcegraph.GetFileOptions{
				FileRange: sourcegraph.FileRange{
					StartLine: int64(def.StartLine),
					EndLine:   int64(def.EndLine),
				},
			},
			NoSrclibAnns: false,
		}
		defEntry, err = cl.RepoTree.Get(r.Context(), &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &opt})
		if err != nil {
			return err
		}
		defAnns, err := cl.Annotations.List(r.Context(), &sourcegraph.AnnotationsListOptions{
			Entry:        entrySpec,
			Range:        &opt.FileRange,
			NoSrclibAnns: opt.NoSrclibAnns,
		})
		if err != nil {
			return err
		}
		defSnippet = &snippet.Snippet{
			StartByte:   defEntry.FileRange.StartByte,
			Code:        defEntry.ContentsString,
			Annotations: defAnns,
		}
		viewDefURL = approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, int(def.StartLine)).String()

		defFileURL = approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, 0).String()

		// fetch example
		refs, err := cl.Defs.ListRefs(r.Context(), &sourcegraph.DefsListRefsOp{
			Def: defSpec,
			Opt: &sourcegraph.DefListRefsOptions{ListOptions: sourcegraph.ListOptions{PerPage: 3}},
		})
		if err != nil {
			return err
		}
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
				NoSrclibAnns: false,
			}
			refRepo, err := cl.Repos.Resolve(r.Context(), &sourcegraph.RepoResolveOp{Path: ref.Repo})
			if err != nil {
				return err
			}
			refEntrySpec := sourcegraph.TreeEntrySpec{
				RepoRev: sourcegraph.RepoRevSpec{Repo: refRepo.Repo, CommitID: ref.CommitID},
				Path:    ref.File,
			}
			refEntry, err := cl.RepoTree.Get(r.Context(), &sourcegraph.RepoTreeGetOp{Entry: refEntrySpec, Opt: opt})
			if err != nil {
				return fmt.Errorf("could not get ref tree: %s", err)
			}
			refAnns, err := cl.Annotations.List(r.Context(), &sourcegraph.AnnotationsListOptions{
				Entry: refEntrySpec,
				Range: &sourcegraph.FileRange{
					// note(beyang): specify line range here, instead of byte range, because the
					// annotation byte offsets will be relative to the start of the snippet in the
					// former, but relative to the start of the file in the latter. This makes the
					// behavior consistent with the def snippet.
					StartLine: refEntry.FileRange.StartLine,
					EndLine:   refEntry.FileRange.EndLine,
				},
				NoSrclibAnns: opt.NoSrclibAnns,
			})
			if err != nil {
				return err
			}
			refSnippets = append(refSnippets, &snippet.Snippet{
				StartByte:   refEntry.FileRange.StartByte,
				Code:        refEntry.ContentsString,
				Annotations: refAnns,
				SourceURL:   approuter.Rel.URLToBlob(ref.Repo, ref.CommitID, ref.File, int(refEntry.FileRange.StartLine+1)).String(),
			})
		}
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

	return tmpl.Exec(r, w, "deflanding.html", http.StatusOK, nil, &struct {
		tmpl.Common
		Meta                meta
		Description         *pbtypes.HTML
		RefSnippets         []*snippet.Snippet
		ViewDefURL          string
		DefName             string // e.g. "func NewRouter"
		ShortDefName        string // e.g. "NewRouter"
		DefFileURL          string
		DefFileName         string
		TotalRepoReferences int32
		DefSnippet          *snippet.Snippet
		RefLocs             *sourcegraph.RefLocationsList
		TruncatedRefLocs    bool
	}{
		Meta:                *m,
		Description:         def.DocHTML,
		RefSnippets:         refSnippets,
		ViewDefURL:          viewDefURL,
		DefName:             def.FmtStrings.DefKeyword + " " + def.FmtStrings.Name.ScopeQualified,
		ShortDefName:        def.Name,
		DefFileURL:          defFileURL,
		DefFileName:         repo.URI + "/" + def.Def.File,
		TotalRepoReferences: refLocs.TotalRepos,
		DefSnippet:          defSnippet,
		RefLocs:             refLocs,
		TruncatedRefLocs:    refLocs.TotalRepos > int32(len(refLocs.RepoRefs)),
	})
}
