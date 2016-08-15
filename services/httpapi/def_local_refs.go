package httpapi

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

// LocalRefLocationsList lists the locations that reference a def in same repository.
type LocalRefLocationsList struct {
	TotalFiles int
	Files      []*sourcegraph.DefFileRef
}

type DefFileRefs []*sourcegraph.DefFileRef

func (list DefFileRefs) Len() int {
	return len(list)
}

func (list DefFileRefs) Less(i, j int) bool {
	return list[i].Count > list[j].Count
}

func (list DefFileRefs) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func serveDefLocalRefLocations(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.DefListRefLocationsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	dc, repo, err := handlerutil.GetDefCommon(ctx, mux.Vars(r), nil)
	if err != nil {
		return err
	}

	if opt.ListOptions.PerPage == 0 && opt.ListOptions.PageOrDefault() == 1 {
		opt.ListOptions.PerPage = 1000
	}

	if feature.IsUniverseRepo(repo.URI) {
		repo, repoRev, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
		if err != nil {
			return err
		}

		file := r.URL.Query().Get("file")

		line, err := strconv.Atoi(r.URL.Query().Get("line"))
		if err != nil {
			return err
		}

		character, err := strconv.Atoi(r.URL.Query().Get("character"))
		if err != nil {
			return err
		}

		localRefs, err := langp.DefaultClient.LocalRefs(&langp.Position{
			Repo:      repo.URI,
			Commit:    repoRev.CommitID,
			File:      file,
			Line:      line,
			Character: character,
		})
		if err != nil {
			return err
		}

		// TODO: we currently only show files not specific location of references,
		// so need to redesign the response type struct and adjust following code logic.
		fileSet := make(map[string]int32)
		for _, ref := range localRefs.Refs {
			if _, ok := fileSet[ref.File]; ok {
				fileSet[ref.File]++
			} else {
				fileSet[ref.File] = 1
			}
		}

		localRefLocationsList := &LocalRefLocationsList{
			TotalFiles: len(fileSet),
			Files:      make([]*sourcegraph.DefFileRef, 0, len(fileSet)),
		}

		for name, count := range fileSet {
			localRefLocationsList.Files = append(localRefLocationsList.Files, &sourcegraph.DefFileRef{
				Path:  name,
				Count: count,
			})
		}

		sort.Sort(DefFileRefs(localRefLocationsList.Files))

		return writeJSON(w, &localRefLocationsList)
	}

	def := dc.Def
	defSpec := sourcegraph.DefSpec{
		Repo:     repo.ID,
		Unit:     def.Unit,
		UnitType: def.UnitType,
		Path:     def.Path,
	}

	refLocations, err := cl.Defs.ListRefLocations(ctx, &sourcegraph.DefsListRefLocationsOp{
		Def: defSpec,
		Opt: &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, refLocations)
}
