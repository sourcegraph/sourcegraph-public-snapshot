package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func serveRepoHoverInfo(w http.ResponseWriter, r *http.Request) error {
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

	var resp = &struct {
		Title string
		Def   *sourcegraph.Def `json:"def"`
	}{}

	if feature.IsUniverseRepo(repo.URI) {
		hover, err := langp.DefaultClient.Hover(&langp.Position{
			Repo:      repo.URI,
			Commit:    repoRev.CommitID,
			File:      file,
			Line:      line,
			Character: character,
		})
		if err != nil {
			return err
		}
		resp.Title = hover.Title
		resp.Def = &sourcegraph.Def{
			Def: graph.Def{
				DefKey: graph.DefKey{Repo: repo.URI},
			},
			DocHTML: htmlutil.SanitizeForPB(hover.DocHTML),
		}
		return writeJSON(w, resp)
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

	resp.Def, err = cl.Defs.Get(ctx, &sourcegraph.DefsGetOp{
		Def: *defSpec,
		Opt: &sourcegraph.DefGetOptions{
			Doc: true,
		},
	})
	if err != nil {
		return err
	}

	return writeJSON(w, resp)
}
