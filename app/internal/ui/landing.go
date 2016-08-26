package ui

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveDefLanding(w http.ResponseWriter, r *http.Request) error {
	// TODO: load GlobalNav after the fact?

	ctx, cl := handlerutil.Client(r)
	vars := mux.Vars(r)

	repo, repoRev, err := handlerutil.GetRepoAndRev(ctx, vars)
	if err != nil {
		return err
	}

	var def *sourcegraph.Def
	var refLocs *sourcegraph.RefLocationsList
	var defEntry *sourcegraph.TreeEntry
	var defSnippet *app.Snippet
	var refEntries []*sourcegraph.TreeEntry
	var refSnippets []*app.Snippet

	if feature.IsUniverseRepo(repo.URI) {
		// TODO
	}

	if def == nil {
		def, _, err = handlerutil.GetDefCommon(ctx, vars, &sourcegraph.DefGetOptions{Doc: true, ComputeLineRange: true})
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
		const reflocRepoLimit = 2
		refLocs, err = cl.Defs.ListRefLocations(ctx, &sourcegraph.DefsListRefLocationsOp{
			Def: defSpec,
			Opt: &sourcegraph.DefListRefLocationsOptions{
				ListOptions: sourcegraph.ListOptions{
					PerPage: reflocRepoLimit, // NOTE(mate): this has no effect at the moment
				},
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
		refLocs.RepoRefs = refLocs.RepoRefs[0:truncLen]

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
		defEntry, err = cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec, Opt: &opt})
		if err != nil {
			return err
		}
		defAnns, err := cl.Annotations.List(ctx, &sourcegraph.AnnotationsListOptions{
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
			SourceURL:   approuter.Rel.URLToBlob(def.Repo, def.CommitID, def.File, 0).String(),
		}

		// fetch example
		refs, err := cl.Defs.ListRefs(ctx, &sourcegraph.DefsListRefsOp{
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
			refRepo, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: ref.Repo})
			if err != nil {
				return err
			}
			refEntrySpec := sourcegraph.TreeEntrySpec{
				RepoRev: sourcegraph.RepoRevSpec{Repo: refRepo.Repo, CommitID: ref.CommitID},
				Path:    ref.File,
			}
			refEntry, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: refEntrySpec, Opt: opt})
			if err != nil {
				return fmt.Errorf("could not get ref tree: %s", err)
			}
			refEntries = append(refEntries, refEntry)
			refAnns, err := cl.Annotations.List(ctx, &sourcegraph.AnnotationsListOptions{
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
				SourceURL:   approuter.Rel.URLToBlob(ref.Repo, ref.CommitID, ref.File, 0).String(),
			})
		}
	}

	return tmpl.Exec(r, w, "deflanding.html", http.StatusOK, nil, &struct {
		tmpl.Common
		Meta meta

		Repo             *sourcegraph.Repo
		RepoRev          sourcegraph.RepoRevSpec
		Def              *sourcegraph.Def
		DefEntry         *sourcegraph.TreeEntry
		DefSnippet       *app.Snippet
		RefLocs          *sourcegraph.RefLocationsList
		TruncatedRefLocs bool
		RefEntries       []*sourcegraph.TreeEntry
		RefSnippets      []*app.Snippet
	}{
		Meta: meta{
			SEO: true,
		},
		Repo:             repo,
		RepoRev:          repoRev,
		Def:              def,
		DefEntry:         defEntry,
		DefSnippet:       defSnippet,
		RefLocs:          refLocs,
		TruncatedRefLocs: refLocs.TotalRepos > int32(len(refLocs.RepoRefs)),
		RefEntries:       refEntries,
		RefSnippets:      refSnippets,
	})
}
