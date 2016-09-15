package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/universe"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/htmlutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

func serveRepoHoverInfo(w http.ResponseWriter, r *http.Request) error {
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

	var resp = &struct {
		Title      string
		Def        *sourcegraph.Def `json:"def"`
		Unresolved bool
	}{}

	if universe.Enabled(r.Context(), repo.URI) {
		hover, err := langp.DefaultClient.Hover(r.Context(), &langp.Position{
			Repo:      repo.URI,
			Commit:    repoRev.CommitID,
			File:      file,
			Line:      line,
			Character: character,
		})
		universeObserve("Hover", err)
		if err != nil {
			return err
		}

		if hover.Unresolved {
			resp.Unresolved = true
			w.Header().Set("cache-control", "private, max-age=60")
			return writeJSON(w, resp)
		}

		var defKey graph.DefKey
		if hover.DefSpec != nil {
			defKey = graph.DefKey{
				Repo:     hover.DefSpec.Repo,
				CommitID: hover.DefSpec.Commit,
				UnitType: hover.DefSpec.UnitType,
				Unit:     hover.DefSpec.Unit,
				Path:     hover.DefSpec.Path,
			}
		}
		resp.Title = hover.Title
		resp.Def = &sourcegraph.Def{
			Def: graph.Def{
				DefKey: defKey,
			},
			DocHTML: htmlutil.SanitizeForPB(hover.DocHTML),
		}
		w.Header().Set("cache-control", "private, max-age=60")
		return writeJSON(w, resp)
	} else if universe.Shadow(repo.URI) {
		go func() {
			_, err := langp.DefaultClient.Hover(r.Context(), &langp.Position{
				Repo:      repo.URI,
				Commit:    repoRev.CommitID,
				File:      file,
				Line:      line,
				Character: character,
			})
			universeObserve("Hover", err)
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

	resp.Def, err = cl.Defs.Get(r.Context(), &sourcegraph.DefsGetOp{
		Def: *defSpec,
		Opt: &sourcegraph.DefGetOptions{
			Doc: true,
		},
	})
	if err != nil {
		return err
	}

	w.Header().Set("cache-control", "private, max-age=60")
	return writeJSON(w, resp)
}
