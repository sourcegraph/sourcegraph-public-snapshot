package httpapi

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func serveJumpToDef(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

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

	var response = &struct {
		Path string `json:"path"`
	}{}

	if feature.IsUniverseRepo(repo.URI) {
		defSpec, err := langp.DefaultClient.PositionToDefSpec(ctx, &langp.Position{
			Repo:      repo.URI,
			Commit:    repoRev.CommitID,
			File:      file,
			Line:      line,
			Character: character,
		})
		if err != nil {
			return err
		}

		rev := ""
		if repoRev.CommitID != "" {
			rev = "@" + repoRev.CommitID
		}
		response.Path = fmt.Sprintf("/%s%s/-/def/%s", repo.URI, rev, defSpec.DefString())
		return writeJSON(w, response)
	}

	defSpec, err := cl.Annotations.GetDefAtPos(ctx, &sourcegraph.AnnotationsGetDefAtPosOptions{
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
	// We still need the string name (not the UID) of the repository to send back.
	def, err := cl.Defs.Get(ctx,
		&sourcegraph.DefsGetOp{
			Def: sourcegraph.DefSpec{
				Repo:     defSpec.Repo,
				CommitID: defSpec.CommitID,
				UnitType: defSpec.UnitType,
				Unit:     defSpec.Unit,
				Path:     defSpec.Path},
			Opt: nil})
	if err != nil {
		return err
	}

	graphKey := graph.DefKey{Repo: def.Repo, CommitID: def.CommitID, UnitType: def.UnitType, Unit: def.Unit, Path: def.Path}
	response.Path = router.Rel.URLToDefKey(graphKey).String()
	return writeJSON(w, response)
}
