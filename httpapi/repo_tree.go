package httpapi

import (
	"net/http"

	"github.com/sourcegraph/mux"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func serveRepoTree(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.RepoTreeGetOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	ctx, _ := handlerutil.Client(r)
	tc, _, _, err := handlerutil.GetTreeEntryCommon(ctx, mux.Vars(r), &opt)
	if err != nil {
		return err
	}
	return writeJSON(w, tc.Entry)
}
