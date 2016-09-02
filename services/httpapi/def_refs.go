package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/universe"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

// RefLocation represents location in file of a reference to a definition.
type RefLocation struct {
	Repo      string
	CommitID  string
	UnitType  string
	Unit      string
	File      string
	StartLine int
	EndLine   int
}

func serveDefRefs(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	var tmp struct {
		Repo string
		sourcegraph.DefListRefsOptions
	}
	if err := schemaDecoder.Decode(&tmp, r.URL.Query()); err != nil {
		return err
	}
	opt := tmp.DefListRefsOptions
	if tmp.Repo != "" {
		var err error
		opt.Repo, err = getRepoID(r.Context(), repoIDOrPath(tmp.Repo))
		if err != nil {
			return err
		}
	}

	dc, repo, err := handlerutil.GetDefCommon(r.Context(), mux.Vars(r), nil)
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

		res, err := cl.Repos.ResolveRev(r.Context(), &sourcegraph.ReposResolveRevOp{Repo: opt.Repo, Rev: ""})
		if err != nil {
			return err
		}

		dataVersion, err := cl.Repos.GetSrclibDataVersionForPath(r.Context(), &sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{Repo: opt.Repo, CommitID: res.CommitID},
			Path:    path,
		})
		if err != nil {
			return err
		}
		opt.CommitID = dataVersion.CommitID
	} else {
		opt.CommitID = def.CommitID
	}

	if universe.Enabled(r.Context(), repo.URI) {
		refs, err := langp.DefaultClient.DefSpecRefs(r.Context(), &langp.DefSpec{
			Repo:     tmp.Repo,
			Commit:   opt.CommitID,
			Unit:     def.Unit,
			UnitType: def.UnitType,
			Path:     def.Path,
		})
		if err != nil {
			return err
		}

		refLocations := make([]*RefLocation, 0, len(refs.Refs))
		for _, ref := range refs.Refs {
			// TODO: investigate whether or not we can remove Def* fields,
			// they seems not been used in frontend.
			refLocations = append(refLocations, &RefLocation{
				Repo:      ref.Repo,
				CommitID:  ref.Commit,
				File:      ref.File,
				StartLine: ref.StartLine,
				EndLine:   ref.EndLine,
			})
		}
		return writeJSON(w, refLocations)
	}

	refs, err := cl.Defs.ListRefs(r.Context(), &sourcegraph.DefsListRefsOp{
		Def: defSpec,
		Opt: &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, refs.Refs)
}

func serveDefRefLocations(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	var opt sourcegraph.DefListRefLocationsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dc, repo, err := handlerutil.GetDefCommon(r.Context(), mux.Vars(r), nil)
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

	refLocations, err := cl.Defs.ListRefLocations(r.Context(), &sourcegraph.DefsListRefLocationsOp{
		Def: defSpec,
		Opt: &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, refLocations)
}

func serveDefExamples(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	var opt sourcegraph.DefsListExamplesOp
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dc, repo, err := handlerutil.GetDefCommon(r.Context(), mux.Vars(r), nil)
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
	opt.Def = defSpec

	opt.ListOptions.PerPage = 3
	opt.ListOptions.Page = 1

	refLocations, err := cl.Defs.ListExamples(r.Context(), &opt)
	if err != nil {
		return err
	}

	return writeJSON(w, refLocations)
}
