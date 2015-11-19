package app

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sourcegraph/mux"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"

	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepoDiscussions(w http.ResponseWriter, r *http.Request) error {
	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}

	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)
	v := mux.Vars(r)
	latest, err := cl.Discussions.List(ctx, &sourcegraph.DiscussionListOp{
		Repo:  sourcegraph.RepoSpec{v["Repo"]},
		Order: sourcegraph.DiscussionListOrder_Date,
	})
	if err != nil {
		return err
	}
	trending, err := cl.Discussions.List(ctx, &sourcegraph.DiscussionListOp{
		Repo:  sourcegraph.RepoSpec{v["Repo"]},
		Order: sourcegraph.DiscussionListOrder_Top,
	})
	if err != nil {
		return err
	}
	return tmpl.Exec(r, w, "repo/discussions.html", http.StatusOK, nil, &struct {
		*handlerutil.RepoCommon
		*handlerutil.RepoRevCommon
		tmpl.Common
		Latest   []*sourcegraph.Discussion
		Trending []*sourcegraph.Discussion
	}{
		RepoCommon:    rc,
		RepoRevCommon: vc,
		Latest:        latest.Discussions,
		Trending:      trending.Discussions,
	})
}

func serveRepoDiscussion(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		return serveRepoDiscussionUpdate(w, r)
	}
	return serveRepoDiscussionView(w, r)
}

func serveRepoDiscussionView(w http.ResponseWriter, r *http.Request) error {
	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)
	v := mux.Vars(r)
	id, err := strconv.ParseInt(v["ID"], 10, 64)
	if err != nil {
		return err
	}
	op := sourcegraph.DiscussionSpec{sourcegraph.RepoSpec{URI: v["Repo"]}, id}
	d, err := cl.Discussions.Get(ctx, &op)
	if err != nil {
		return err
	}
	related, err := cl.Discussions.List(ctx, &sourcegraph.DiscussionListOp{
		DefKey: d.DefKey,
		Order:  sourcegraph.DiscussionListOrder_Top,
	})
	if err != nil {
		return err
	}
	return tmpl.Exec(r, w, "repo/discussion.html", http.StatusOK, nil, &struct {
		*handlerutil.RepoCommon
		*handlerutil.RepoRevCommon
		tmpl.Common
		Discussion *sourcegraph.Discussion
		Related    []*sourcegraph.Discussion
	}{
		RepoCommon:    rc,
		RepoRevCommon: vc,
		Discussion:    d,
		Related:       related.Discussions,
	})
}

func serveRepoDiscussionUpdate(w http.ResponseWriter, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	frm := r.Form.Get
	id, err := strconv.ParseInt(frm("discussion_id"), 10, 64)
	if err != nil {
		return err
	}
	var dk graph.DefKey
	if err := json.Unmarshal([]byte(frm("def_key")), &dk); err != nil {
		return err
	}
	u := handlerutil.UserFromRequest(r)
	if u == nil {
		return &errcode.HTTPErr{Status: http.StatusUnauthorized}
	}
	c := &sourcegraph.DiscussionComment{
		Body:   frm("body"),
		Author: *u,
		DefKey: dk,
	}
	op := sourcegraph.DiscussionCommentCreateOp{id, c}
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
	_, err = cl.Discussions.CreateComment(ctx, &op)
	if err != nil {
		return err
	}
	return serveRepoDiscussionView(w, r)
}
