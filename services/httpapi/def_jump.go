package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/universe"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func serveJumpToDef(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	repo, repoRev, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
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

	var response graph.DefKey

	if universe.EnabledFile(file) {
		defRange, err := langp.DefaultClient.Definition(r.Context(), &langp.Position{
			Repo:      repo.URI,
			Commit:    repoRev.CommitID,
			File:      file,
			Line:      line,
			Character: character,
		})
		universeObserve("Definition", err)
		if err != nil {
			return err
		}
		if defRange.Empty() {
			response.Path = ""
		} else if defRange.File == "." {
			// Special case the top level directory
			response.Path = router.Rel.URLToRepoTreeEntry(defRange.Repo, defRange.Commit, "").String()
		} else {
			// We increment the line number by 1 because the blob view is not zero-indexed.
			response.Path = router.Rel.URLToBlobRange(defRange.Repo, defRange.Commit, defRange.File, defRange.StartLine+1, defRange.EndLine+1, defRange.StartCharacter+1, defRange.EndCharacter+2).String()
		}
		w.Header().Set("cache-control", "private, max-age=60")
		return writeJSON(w, response)
	} else if universe.Shadow(repo.URI) {
		go func() {
			_, err := langp.DefaultClient.Definition(r.Context(), &langp.Position{
				Repo:      repo.URI,
				Commit:    repoRev.CommitID,
				File:      file,
				Line:      line,
				Character: character,
			})
			universeObserve("Definition", err)
		}()
	}

	defSpec, err := cl.Annotations.GetDefAtPos(r.Context(), &sourcegraph.AnnotationsGetDefAtPosOptions{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: repoRev,
			Path:    file,
		},
		Line:      uint32(line),
		Character: uint32(character),
	})
	if err != nil {
		return err
	}
	def, err := cl.Defs.Get(r.Context(),
		&sourcegraph.DefsGetOp{
			Def: sourcegraph.DefSpec{
				Repo:     defSpec.Repo,
				CommitID: defSpec.CommitID,
				UnitType: defSpec.UnitType,
				Unit:     defSpec.Unit,
				Path:     defSpec.Path},
			Opt: &sourcegraph.DefGetOptions{ComputeLineRange: true}})
	if err != nil {
		return err
	}

	// Don't add an ugly commit ID to the URL in the address bar if
	// the user is on a named branch or default branch. Also, never
	// add a rev if the jump is cross-repo.
	var rev string
	if v := routevar.ToRepoRev(mux.Vars(r)); def.Repo == v.Repo {
		rev = v.Rev
	}

	response.Path = router.Rel.URLToBlob(def.Repo, rev, def.File, int(def.StartLine)).String()
	w.Header().Set("cache-control", "private, max-age=60")
	return writeJSON(w, response)
}
