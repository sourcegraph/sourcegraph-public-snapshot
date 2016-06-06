package httpapi

import (
	"net/http"
	"sort"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveDefRefs(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var tmp struct {
		Repo repoIDOrPath
		sourcegraph.DefListRefsOptions
	}
	if err := schemaDecoder.Decode(&tmp, r.URL.Query()); err != nil {
		return err
	}
	opt := tmp.DefListRefsOptions
	if tmp.Repo != "" {
		var err error
		opt.Repo, err = getRepoID(ctx, tmp.Repo)
		if err != nil {
			return err
		}
	}

	dc, repo, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), nil)
	if err != nil {
		return err
	}

	def := dc.Def
	defSpec := sourcegraph.DefSpec{
		Repo:     repo.ID,
		CommitID: def.CommitID,
		Unit:     def.Unit,
		UnitType: def.UnitType,
		Path:     def.Path,
	}

	if opt.ListOptions.PerPage == 0 && opt.ListOptions.PageOrDefault() == 1 {
		opt.ListOptions.PerPage = 10000
	}
	if opt.Repo == 0 {
		opt.Repo = defSpec.Repo
	}
	// Restrict search for external repo refs to the last built commit on the default branch
	// of the external repo.
	if opt.Repo != defSpec.Repo {
		var path string
		// If the ref search is restricted to one file of the repo, make sure we have build
		// data for that file. Otherwise, use the most recent commit that is built.
		if len(opt.Files) == 1 {
			path = opt.Files[0]
		}

		res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: opt.Repo, Rev: ""})
		if err != nil {
			return err
		}

		dataVersion, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{Repo: opt.Repo, CommitID: res.CommitID},
			Path:    path,
		})
		if err != nil {
			return err
		}
		opt.CommitID = dataVersion.CommitID
	}

	refs, err := cl.Defs.ListRefs(ctx, &sourcegraph.DefsListRefsOp{
		Def: defSpec,
		Opt: &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, refs.Refs)
}

func serveDefRefLocations(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.DefListRefLocationsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dc, repo, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), nil)
	if err != nil {
		return err
	}

	def := dc.Def
	defSpec := sourcegraph.DefSpec{
		Repo:     repo.ID,
		Unit:     def.Unit,
		UnitType: def.UnitType,
		Path:     def.Path,
	}

	if opt.ListOptions.PerPage == 0 && opt.ListOptions.PageOrDefault() == 1 {
		opt.ListOptions.PerPage = 1000
	}

	refLocations, err := cl.Defs.ListRefLocations(ctx, &sourcegraph.DefsListRefLocationsOp{
		Def: defSpec,
		Opt: &opt,
	})
	if err != nil {
		return err
	}

	// TEMPORARY FIX: refs are not available for the current repo because the HEAD commit
	// of the default branch of the repo hasn't been built yet after the switch to pgsql
	// global refs store. Fallback to fetching local repo refs from graph store.
	//
	// TODO(slimsag): remove this kludge after migration is complete for all existing repos.
	containsDefRepo := false
	for _, refs := range refLocations.RepoRefs {
		if refs.Repo == def.Repo {
			containsDefRepo = true
			break
		}
	}
	if (len(refLocations.RepoRefs) == 0 || !containsDefRepo) && opt.PageOrDefault() == 1 {
		log15.Debug("Missing local refs on DefInfo", "def", defSpec.String(), "lenRepoRefs", len(refLocations.RepoRefs))
		// Scope the local repo ref search to the def's commit ID.
		defSpec.CommitID = def.CommitID
		refs, err := cl.Defs.ListRefs(ctx, &sourcegraph.DefsListRefsOp{
			Def: defSpec,
			Opt: &sourcegraph.DefListRefsOptions{
				Repo:        defSpec.Repo,
				ListOptions: opt.ListOptions,
			},
		})
		if err != nil {
			return err
		}

		refsPerFile := make(map[string]int32)
		totalCount := int32(0)
		for _, ref := range refs.Refs {
			refsPerFile[ref.File]++
			totalCount++
		}

		if totalCount > 0 {
			fl := sortByRefCount(refsPerFile)

			localRefs := &sourcegraph.DefRepoRef{
				Repo:  repo.URI,
				Count: totalCount,
				Files: fl,
			}

			refLocations.RepoRefs = append(refLocations.RepoRefs, localRefs)
			lastIdx := len(refLocations.RepoRefs) - 1
			refLocations.RepoRefs[0], refLocations.RepoRefs[lastIdx] = refLocations.RepoRefs[lastIdx], refLocations.RepoRefs[0]
		}
	}

	return writeJSON(w, refLocations)
}

type fileList []*sourcegraph.DefFileRef

func (f fileList) Len() int           { return len(f) }
func (f fileList) Less(i, j int) bool { return f[i].Count < f[j].Count }
func (f fileList) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func sortByRefCount(refsPerFile map[string]int32) fileList {
	fl := make(fileList, len(refsPerFile))
	i := 0
	for k, v := range refsPerFile {
		fl[i] = &sourcegraph.DefFileRef{Path: k, Count: v}
		i++
	}
	sort.Sort(sort.Reverse(fl))
	return fl
}
