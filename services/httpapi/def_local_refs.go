package httpapi

import (
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/universe"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

// LocalRefLocationsList lists the locations that reference a def in same repository.
type LocalRefLocationsList struct {
	TotalFiles int
	Files      []*sourcegraph.DefFileRef
	Locs       []lsp.Location
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
	cl := handlerutil.Client(r)

	var opt sourcegraph.DefListRefLocationsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	wantUniverseResults := r.URL.Query().Get("file") != "" && r.URL.Query().Get("line") != "" && r.URL.Query().Get("character") != ""

	dc, repo, err := handlerutil.GetDefCommon(r.Context(), mux.Vars(r), nil)
	if err != nil && !wantUniverseResults {
		// If we want Universe results, we don't need the srclib def to exist.
		return err
	}

	if opt.ListOptions.PerPage == 0 && opt.ListOptions.PageOrDefault() == 1 {
		opt.ListOptions.PerPage = 1000
	}

	if wantUniverseResults {
		localRefLocationsList, err := universeDefLocalRefLocations(r)
		if err != nil {
			return err
		}
		return writeJSON(w, &localRefLocationsList)
	} else if universe.Shadow(repo.URI) {
		go universeDefLocalRefLocations(r)
	}
	if wantUniverseResults {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("def not universe-enabled")}
	}

	def := dc.Def
	defSpec := sourcegraph.DefSpec{
		Repo:     repo.ID,
		Unit:     def.Unit,
		UnitType: def.UnitType,
		Path:     def.Path,
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

func universeDefLocalRefLocations(r *http.Request) (*LocalRefLocationsList, error) {
	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return nil, err
	}

	file := r.URL.Query().Get("file")

	line, err := strconv.Atoi(r.URL.Query().Get("line"))
	if err != nil {
		return nil, err
	}

	character, err := strconv.Atoi(r.URL.Query().Get("character"))
	if err != nil {
		return nil, err
	}

	localRefs, err := langp.DefaultClient.LocalRefs(r.Context(), &langp.Position{
		Repo:      repo.URI,
		Commit:    repoRev.CommitID,
		File:      file,
		Line:      line,
		Character: character,
	})
	universeObserve("LocalRefs", err)
	if err != nil {
		return nil, err
	}

	// TODO: we currently only show files not specific location of references,
	// so need to redesign the response type struct and adjust following code logic.
	var locs []lsp.Location
	fileSet := make(map[string]int32)
	for _, ref := range localRefs.Refs {
		if _, ok := fileSet[ref.File]; ok {
			fileSet[ref.File]++
		} else {
			fileSet[ref.File] = 1
		}
		locs = append(locs, lsp.Location{
			URI: makeLSPURI(ref.Repo, "", ref.File),
			Range: lsp.Range{
				Start: lsp.Position{Line: ref.StartLine, Character: ref.StartCharacter},
				End:   lsp.Position{Line: ref.EndLine, Character: ref.EndCharacter + 1},
			},
		})
	}

	localRefLocationsList := &LocalRefLocationsList{
		TotalFiles: len(fileSet),
		Files:      make([]*sourcegraph.DefFileRef, 0, len(fileSet)),
		Locs:       locs,
	}

	for name, count := range fileSet {
		localRefLocationsList.Files = append(localRefLocationsList.Files, &sourcegraph.DefFileRef{
			Path:  name,
			Count: count,
		})
	}

	sort.Sort(DefFileRefs(localRefLocationsList.Files))
	return localRefLocationsList, nil
}
