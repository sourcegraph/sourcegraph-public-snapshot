package app

import (
	"net/http"

	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveUser(w http.ResponseWriter, r *http.Request) error {
	if r.URL.Path == "/explore" {
		// Redirect the old /explore path to the homepage for now.
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return nil
	}

	var opt sourcegraph.RepoListOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	p, spec, err := getUser(ctx, r)
	if err != nil {
		return err
	}

	opt.Owner = spec.Login
	opt.Sort = "pushed"
	opt.Direction = "desc"
	if opt.PerPage == 0 {
		opt.PerPage = 50
	}

	repos, err := apiclient.Repos.List(ctx, &opt)
	if err != nil {
		return err
	}

	pg, err := paginate(opt /* TODO */, 0)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "user/owned_repos.html", http.StatusOK, nil, &struct {
		User        *sourcegraph.User
		Repos       []*sourcegraph.Repo
		PageLinks   []pageLink
		RobotsIndex bool
		tmpl.Common
	}{
		User:        p,
		Repos:       repos.Repos,
		PageLinks:   pg,
		RobotsIndex: true,
	})
}

func serveUserOrgs(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.ListOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	p, spec, err := getUser(ctx, r)
	if err != nil {
		return err
	}

	orgs, err := apiclient.Orgs.List(ctx, &sourcegraph.OrgsListOp{Member: *spec, ListOptions: opt})
	if errcode.GRPC(err) == codes.Unimplemented {
		orgs = &sourcegraph.OrgList{} // ignore error
	} else if err != nil {
		return err
	}

	// TODO(sqs): implement pagination in API for org memberships
	// pg, err := paginate(opt, p.Stat[client.StatOrgMemberships])
	// if err != nil {
	// 	return err
	// }

	return tmpl.Exec(r, w, "user/orgs.html", http.StatusOK, nil, &struct {
		User      *sourcegraph.User
		Orgs      []*sourcegraph.Org
		PageLinks []pageLink
		tmpl.Common
	}{
		User: p,
		Orgs: orgs.Orgs,
	})
}
