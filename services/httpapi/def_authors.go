package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func serveDefAuthors(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.DefListAuthorsOptions
	if err := schemaDecoder.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	defSpec, err := resolveDef(ctx, routevar.ToDefAtRev(mux.Vars(r)))
	if err != nil {
		return err
	}

	authors, err := cl.Defs.ListAuthors(ctx, &sourcegraph.DefsListAuthorsOp{
		Def: *defSpec,
		Opt: &opt,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, authors)
}
