package ui

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func serveChangesetCreate(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	var newChangeset *sourcegraph.Changeset
	if err := json.NewDecoder(r.Body).Decode(&newChangeset); err != nil {
		return err
	}
	defer r.Body.Close()
	uri := mux.Vars(r)["Repo"]

	user := handlerutil.UserFromRequest(r)
	if user == nil {
		return &handlerutil.HTTPErr{Status: http.StatusUnauthorized}
	}

	newChangeset.Author = user.Spec()
	cs, err := cl.Changesets.Create(ctx, &sourcegraph.ChangesetCreateOp{
		Repo:      sourcegraph.RepoSpec{URI: uri},
		Changeset: newChangeset,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(struct {
		Repo string
		ID   int64
	}{uri, cs.ID})
}

func serveChangesetSubmitReview(w http.ResponseWriter, r *http.Request) error {
	v := mux.Vars(r)
	id, err := strconv.ParseInt(v["ID"], 10, 64)
	if err != nil {
		return err
	}
	uri := v["Repo"]
	newReview := &sourcegraph.ChangesetReview{}
	if err := json.NewDecoder(r.Body).Decode(newReview); err != nil {
		return err
	}
	defer r.Body.Close()

	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	user := handlerutil.UserFromRequest(r)
	if user == nil {
		return &handlerutil.HTTPErr{Status: http.StatusUnauthorized}
	}

	newReview.Author = user.Spec()
	review, err := cl.Changesets.CreateReview(ctx, &sourcegraph.ChangesetCreateReviewOp{
		Repo:        sourcegraph.RepoSpec{URI: uri},
		ChangesetID: id,
		Review:      newReview,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(review)
}

func serveChangesetUpdate(w http.ResponseWriter, r *http.Request) error {
	v := mux.Vars(r)
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	id, err := strconv.ParseInt(v["ID"], 10, 64)
	if err != nil {
		return err
	}
	var op sourcegraph.ChangesetUpdateOp
	if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
		return err
	}
	defer r.Body.Close()
	op.ID = id
	op.Repo = sourcegraph.RepoSpec{URI: v["Repo"]}

	user := handlerutil.UserFromRequest(r)
	if user == nil {
		return &handlerutil.HTTPErr{Status: http.StatusUnauthorized}
	}

	op.Author = user.Spec()
	result, err := cl.Changesets.Update(ctx, &op)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(result)
}
