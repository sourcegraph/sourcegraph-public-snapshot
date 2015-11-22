package app

import (
	"net/http"
	"net/url"
	"os"

	"code.google.com/p/rog-go/parallel"

	"sync"

	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := sourcegraph.NewClientFromContext(ctx)

	conf, err := cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	rootURL, err := url.Parse(conf.FederationRootURL)
	if err != nil {
		return err
	}

	var listOpts sourcegraph.ListOptions
	if err := schemautil.Decode(&listOpts, r.URL.Query()); err != nil {
		return err
	}

	if listOpts.PerPage == 0 {
		listOpts.PerPage = 50
	}

	repos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: listOpts,
	})
	if err != nil {
		return err
	}
	var (
		users   []string
		usersMu sync.Mutex
	)
	userPerms, err := cl.RegisteredClients.ListUserPermissions(ctx, &sourcegraph.RegisteredClientSpec{})
	if err != nil && grpc.Code(err) != codes.PermissionDenied && grpc.Code(err) != codes.Unauthenticated {
		return err
	}
	if err == nil { // current user is admin of the instance
		par := parallel.NewRun(10)
		for _, perms_ := range userPerms.UserPermissions {
			perms := perms_
			par.Do(func() error {
				user, err := cl.Users.Get(ctx, &sourcegraph.UserSpec{UID: perms.UID})
				if err != nil {
					return err
				}
				usersMu.Lock()
				users = append(users, user.Login)
				usersMu.Unlock()
				return nil
			})
		}
		if err := par.Wait(); err != nil {
			return err
		}
	}

	return tmpl.Exec(r, w, "home/dashboard.html", http.StatusOK, nil, &struct {
		Repos  []*sourcegraph.Repo
		SGPath string
		Users  []string

		RootURL *url.URL

		tmpl.Common
	}{
		Repos:  repos.Repos,
		SGPath: os.Getenv("SGPATH"),
		Users:  users,

		RootURL: rootURL,
	})
}
