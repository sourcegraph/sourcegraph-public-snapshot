package ui

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
)

type defDescr struct {
	Def       *sourcegraph.Def
	RefCount  int32
	LandURL   string
	SourceURL string
}

func serveRepoLanding(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)
	vars := mux.Vars(r)

	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), vars)
	if err != nil {
		return err
	}

	// terminate early on non-Go repos
	if repo.Language != "Go" {
		http.Error(w, "404 - Page not found. (No landing page for non-Go repo.)", http.StatusNotFound)
		return nil
	}

	repoURL := approuter.Rel.URLToRepoRev(repo.URI, repoRev.CommitID).String()

	results, err := cl.Search.Search(r.Context(), &sourcegraph.SearchOp{
		Opt: &sourcegraph.SearchOptions{
			Repos:        []int32{repo.ID},
			Languages:    []string{"Go"},
			IncludeRepos: false,
			ListOptions:  sourcegraph.ListOptions{PerPage: 20},
		},
	})
	if err != nil {
		return err
	}

	var defDescrs []defDescr
	for _, defResult := range results.DefResults {
		def := &defResult.Def

		if !def.Exported {
			continue
		}

		if def.Kind == "package" {
			continue
		}

		htmlutil.ComputeDocHTML(def)

		defDescrs = append(defDescrs, defDescr{
			Def:       def,
			RefCount:  defResult.RefCount,
			LandURL:   approuter.Rel.DefKeyToLandURL(def.DefKey).String(),
			SourceURL: approuter.Rel.URLToDefKey(def.DefKey).String(),
		})
	}

	searchIconSVG := `<svg width="16px" viewBox="0 0 17 17" style="top: 11px; left: 10px; vertical-align: middle;"><path fill="rgb(119, 147, 174)" d="M10.6 0c1.2 0 2.2.3 3.2 1 1 .4 1.8 1.2 2.3 2.2.7 1 1 2 1 3.2 0 1-.3 2.2-1 3.2-.4 1-1.2 1.7-2.2 2.3-1 .5-2 .8-3.2.8-1.2 0-2.4-.4-3.4-1l-4.8 4.8c-.3.3-.6.4-1 .4s-.7 0-1-.4c-.3-.3-.4-.6-.4-1s0-.7.4-1l4.8-4.8c-.6-1-1-2.2-1-3.4 0-1.2.3-2.2 1-3.2.5-1 1.2-1.8 2.2-2.3 1-.7 2-1 3.2-1zm0 10.6c.6 0 1 0 1.7-.3.5-.2 1-.5 1.3-1 .4-.3.7-.8 1-1.3.2-.5.3-1 .3-1.6 0-.6-.2-1-.5-1.7-.2-.5-.5-1-1-1.3-.3-.4-.7-.7-1.2-1-.6-.2-1-.3-1.7-.3-.6 0-1 .2-1.6.5-.5.2-1 .5-1.4 1-.4.3-.7.7-1 1.2l-.2 1.7c0 .6 0 1 .3 1.6.2.5.5 1 1 1.4.3.4.8.7 1.3 1l1.6.2z"></path></svg>`

	return tmpl.Exec(r, w, "repolanding.html", http.StatusOK, nil, &struct {
		tmpl.Common
		Meta meta

		MetaTitle     string
		MetaDescr     string
		SearchIconSVG string
		Repo          *sourcegraph.Repo
		RepoRev       sourcegraph.RepoRevSpec
		RepoURL       string
		Defs          []defDescr
	}{
		Meta:          meta{SEO: true},
		MetaTitle:     fmt.Sprintf("%s · Sourcegraph", repo.URI),
		MetaDescr:     fmt.Sprintf("Top definitions from %s with type signature, documentation, links to source and usage examples", repo.URI),
		SearchIconSVG: searchIconSVG,
		Repo:          repo,
		RepoRev:       repoRev,
		RepoURL:       repoURL,
		Defs:          defDescrs,
	})
}

func serveDefLanding(w http.ResponseWriter, r *http.Request) error {
	// TODO: load GlobalNav after the fact?

	cl := handlerutil.Client(r)
	vars := mux.Vars(r)

	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), vars)
	if err != nil {
		return err
	}

	repoURL := approuter.Rel.URLToRepoRev(repo.URI, repoRev.CommitID).String()
	var fileURL string
	var def *sourcegraph.Def
	var refLocs *sourcegraph.RefLocationsList
	var defEntry *sourcegraph.TreeEntry
	var defSnippet *app.Snippet
	var refEntries []*sourcegraph.TreeEntry
	var refSnippets []*app.Snippet

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
		defSnippet = &app.Snippet{
			StartByte:   defEntry.FileRange.StartByte,
			Code:        defEntry.ContentsString,
			Annotations: defAnns,
			SourceURL:   approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, int(def.StartLine)).String(),
		}

		fileURL = approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, 0).String()

		// fetch example
		refs, err := cl.Defs.ListRefs(r.Context(), &sourcegraph.DefsListRefsOp{
			Def: defSpec,
			Opt: &sourcegraph.DefListRefsOptions{ListOptions: sourcegraph.ListOptions{PerPage: 1}},
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
			refEntries = append(refEntries, refEntry)
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
			refSnippets = append(refSnippets, &app.Snippet{
				StartByte:   refEntry.FileRange.StartByte,
				Code:        refEntry.ContentsString,
				Annotations: refAnns,
				SourceURL:   approuter.Rel.URLToBlob(ref.Repo, ref.CommitID, ref.File, int(refEntry.FileRange.StartLine+1)).String(),
			})
		}
	}

	return tmpl.Exec(r, w, "deflanding.html", http.StatusOK, nil, &struct {
		tmpl.Common
		Meta meta

		Repo             *sourcegraph.Repo
		RepoRev          sourcegraph.RepoRevSpec
		RepoURL          string
		FileURL          string
		Def              *sourcegraph.Def
		DefEntry         *sourcegraph.TreeEntry
		DefSnippet       *app.Snippet
		RefLocs          *sourcegraph.RefLocationsList
		TruncatedRefLocs bool
		RefEntries       []*sourcegraph.TreeEntry
		RefSnippets      []*app.Snippet
	}{
		Meta:             meta{SEO: true},
		Repo:             repo,
		RepoRev:          repoRev,
		RepoURL:          repoURL,
		FileURL:          fileURL,
		Def:              def,
		DefEntry:         defEntry,
		DefSnippet:       defSnippet,
		RefLocs:          refLocs,
		TruncatedRefLocs: refLocs.TotalRepos > int32(len(refLocs.RepoRefs)),
		RefEntries:       refEntries,
		RefSnippets:      refSnippets,
	})
}
