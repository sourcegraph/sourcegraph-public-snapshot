package app

import (
	"net/http"
	"os"

	"code.google.com/p/rog-go/parallel"

	"sync"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func serveHomeDashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := sourcegraph.NewClientFromContext(ctx)

	var listOpts sourcegraph.ListOptions
	if err := schemautil.Decode(&listOpts, r.URL.Query()); err != nil {
		return err
	}

	if listOpts.PerPage == 0 {
		listOpts.PerPage = 50
	}

	repos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{ListOptions: listOpts})
	if err != nil {
		return err
	}
	var template string
	var (
		users   []string
		usersMu sync.Mutex
	)
	if len(repos.Repos) > 0 {
		userPerms, err := cl.RegisteredClients.ListUserPermissions(ctx, &sourcegraph.RegisteredClientSpec{})
		if err != nil && grpc.Code(err) != codes.PermissionDenied {
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
		template = "home/dashboard.html"
	} else {
		template = "home/new.html"
	}

	return tmpl.Exec(r, w, template, http.StatusOK, nil, &struct {
		Repos  []*sourcegraph.Repo
		SGPath string
		Users  []string
		tmpl.Common
	}{
		Repos:  repos.Repos,
		SGPath: os.Getenv("SGPATH"),
		Users:  users,
	})
}
