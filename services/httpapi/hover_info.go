package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveRepoHoverInfo(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	_, repoRev, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
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

	def, err := cl.Defs.Get(ctx, &sourcegraph.DefsGetOp{
		Def: *defSpec,
		Opt: &sourcegraph.DefGetOptions{
			Doc: true,
		},
	})
	if err != nil {
		return err
	}

	return writeJSON(w, struct {
		Def *sourcegraph.Def `json:"def"`
	}{
		Def: def,
	})
}
