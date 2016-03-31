package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveRepoBuildsCreate(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var op sourcegraph.BuildsCreateOp
	err := json.NewDecoder(r.Body).Decode(&op)
	if err != nil {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: err}
	}

	_, repoSpec, err := handlerutil.GetRepo(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	op.Repo = repoSpec
	build, err := cl.Builds.Create(ctx, &op)
	if err != nil {
		return err
	}
	return writeJSON(w, build)
}
