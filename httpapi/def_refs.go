package httpapi

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveDefRefs(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.DefListRefsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dc, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), nil)
	if err != nil {
		return err
	}

	def := dc.Def
	defSpec := sourcegraph.DefSpec{
		Repo:     def.Repo,
		CommitID: def.CommitID,
		Unit:     def.Unit,
		UnitType: def.UnitType,
		Path:     def.Path,
	}

	if opt.ListOptions.PerPage == 0 && opt.ListOptions.PageOrDefault() == 1 {
		opt.ListOptions.PerPage = 10000
	}
	if opt.Repo == "" {
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
		dataVersion, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: opt.Repo}},
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

	return json.NewEncoder(w).Encode(refs.Refs)
}

func serveDefRefLocations(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.DefListRefLocationsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dc, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), nil)
	if err != nil {
		return err
	}

	def := dc.Def
	defSpec := sourcegraph.DefSpec{
		Repo:     def.Repo,
		Unit:     def.Unit,
		UnitType: def.UnitType,
		Path:     def.Path,
	}

	if opt.ListOptions.PerPage == 0 && opt.ListOptions.PageOrDefault() == 1 {
		if opt.ReposOnly {
			opt.ListOptions.PerPage = 333
		} else {
			opt.ListOptions.PerPage = 1000
		}
	}

	refLocations, err := cl.Defs.ListRefLocations(ctx, &sourcegraph.DefsListRefLocationsOp{
		Def: defSpec,
		Opt: &opt,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(refLocations)
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
