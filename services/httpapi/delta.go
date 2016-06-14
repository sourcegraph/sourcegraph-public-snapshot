package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func serveDeltaFiles(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.DeltaListFilesOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	delta := routevar.ToDelta(mux.Vars(r))
	baseRepoRev, err := resolveLocalRepoRev(ctx, delta.Base)
	if err != nil {
		return err
	}
	headRepoRev, err := resolveLocalRepoRev(ctx, delta.Head)
	if err != nil {
		return err
	}

	files, err := cl.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{
		Ds:  sourcegraph.DeltaSpec{Base: *baseRepoRev, Head: *headRepoRev},
		Opt: &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, files)
}
