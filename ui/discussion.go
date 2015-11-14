package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveDiscussion(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	v := mux.Vars(r)
	id, err := strconv.ParseInt(v["ID"], 10, 64)
	if err != nil {
		return err
	}

	op := sourcegraph.DiscussionSpec{
		Repo: sourcegraph.RepoSpec{URI: v["Repo"]},
		ID:   id,
	}
	d, err := cl.Discussions.Get(ctx, &op)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(d)
}

func serveDiscussionListDef(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	//repo := sourcegraph.RepoSpec{URI: vars["Repo"]}
	defKey := graph.DefKey{
		Repo:     vars["Repo"],
		UnitType: vars["UnitType"],
		Path:     vars["Path"],
		Unit:     vars["Unit"],
	}
	op := sourcegraph.DiscussionListOp{
		DefKey: defKey,
	}
	return serveDiscussionList(w, r, op)
}

func serveDiscussionListRepo(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	op := sourcegraph.DiscussionListOp{
		Repo: sourcegraph.RepoSpec{URI: vars["Repo"]},
	}
	return serveDiscussionList(w, r, op)
}

func serveDiscussionList(w http.ResponseWriter, r *http.Request, op sourcegraph.DiscussionListOp) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	r.ParseForm()
	order := sourcegraph.DiscussionListOrder_Top
	if o, ok := r.Form["order"]; ok {
		if v, ok := sourcegraph.DiscussionListOrder_value[o[0]]; ok {
			order = sourcegraph.DiscussionListOrder(v)
		} else {
			return fmt.Errorf("Unknown list order %s", o)
		}
	}

	// TODO(keegan) Update op to include list options
	op.Order = order

	dss, err := cl.Discussions.List(ctx, &op)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(dss)
}

func serveDiscussionCreate(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	var d *sourcegraph.Discussion
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		return err
	}
	r.Body.Close()

	vars := mux.Vars(r)
	d.DefKey = graph.DefKey{
		Repo:     vars["Repo"],
		UnitType: vars["UnitType"],
		Path:     vars["Path"],
		Unit:     vars["Unit"],
	}

	user := handlerutil.UserFromRequest(r)
	if user == nil {
		return &handlerutil.HTTPErr{Status: http.StatusUnauthorized}
	}

	d.Author = *user
	d, err := cl.Discussions.Create(ctx, d)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(d)
}

func serveDiscussionCommentCreate(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	var c *sourcegraph.DiscussionComment
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return err
	}
	r.Body.Close()

	vars := mux.Vars(r)
	c.DefKey = graph.DefKey{
		Repo:     vars["Repo"],
		UnitType: vars["UnitType"],
		Path:     vars["Path"],
		Unit:     vars["Unit"],
	}
	id, err := strconv.ParseInt(vars["ID"], 10, 64)
	if err != nil {
		return err
	}

	user := handlerutil.UserFromRequest(r)
	if user == nil {
		return &handlerutil.HTTPErr{Status: http.StatusUnauthorized}
	}
	c.Author = *user

	op := sourcegraph.DiscussionCommentCreateOp{
		DiscussionID: id,
		Comment:      c,
	}
	c, err = cl.Discussions.CreateComment(ctx, &op)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(c)
}
