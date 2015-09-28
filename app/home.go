package app

import (
	"net/http"
	"net/url"
	"os"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/schemautil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sqs/pbtypes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	var listOpts sourcegraph.ListOptions
	if err := schemautil.Decode(&listOpts, r.URL.Query()); err != nil {
		return err
	}

	if listOpts.PerPage == 0 {
		listOpts.PerPage = 50
	}

	repos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{ListOptions: listOpts})
	if err != nil {
		if grpc.Code(err) == codes.Unauthenticated {
			return serveWelcomeInterstitial(w, r)
		}
		return err
	}
	var template string
	if len(repos.Repos) > 0 {
		template = "home/dashboard.html"
	} else {
		template = "home/new.html"
	}

	return tmpl.Exec(r, w, template, http.StatusOK, nil, &struct {
		Repos  []*sourcegraph.Repo
		SGPath string
		tmpl.Common
	}{
		Repos:  repos.Repos,
		SGPath: os.Getenv("SGPATH"),
	})
}

func serveWelcomeInterstitial(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	conf, err := cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	u, err := url.Parse(conf.FederationRootURL)
	if err != nil {
		return err
	}
	return tmpl.Exec(r, w, "home/welcome.html", http.StatusOK, nil, &struct {
		RootHostname string
		tmpl.Common
	}{
		RootHostname: u.Host,
	})
}
