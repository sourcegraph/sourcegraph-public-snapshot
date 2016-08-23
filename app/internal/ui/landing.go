package ui

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
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
	var refEntries []*sourcegraph.TreeEntry

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
		refLocs, err = cl.Defs.ListRefLocations(ctx, &sourcegraph.DefsListRefLocationsOp{Def: defSpec})
		if err != nil {
			return err
		}

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
		}
	}

	return tmpl.Exec(r, w, "deflanding.html", http.StatusOK, nil, &struct {
		tmpl.Common
		Meta meta

		Repo       *sourcegraph.Repo
		RepoRev    sourcegraph.RepoRevSpec
		Def        *sourcegraph.Def
		DefEntry   *sourcegraph.TreeEntry
		RefLocs    *sourcegraph.RefLocationsList
		RefEntries []*sourcegraph.TreeEntry
	}{
		Repo:       repo,
		RepoRev:    repoRev,
		Def:        def,
		DefEntry:   defEntry,
		RefLocs:    refLocs,
		RefEntries: refEntries,
	})
}
