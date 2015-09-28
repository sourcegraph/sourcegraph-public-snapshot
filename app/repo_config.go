package app

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	appauthutil "sourcegraph.com/sourcegraph/sourcegraph/app/internal/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	internal.Handlers[router.RepoEnable] = serveRepoEnable
}

// serveRepoEnable is a simplified repo config update
// endpoint that *only* enables a repo.
func serveRepoEnable(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	if authutil.ActiveFlags.HasUserAccounts() && handlerutil.UserFromRequest(r) == nil {
		return appauthutil.RedirectToLogIn(w, r)
	}

	rc, err := handlerutil.GetRepoCommon(r, &handlerutil.GetRepoCommonOpt{AllowNonEnabledRepos: true})
	if err != nil {
		return err
	}

	var method func(context.Context, *sourcegraph.RepoSpec, ...grpc.CallOption) (*pbtypes.Void, error)
	if enable := r.Method != "DELETE"; enable {
		method = apiclient.Repos.Enable
	} else {
		method = apiclient.Repos.Disable
	}

	repoSpec := rc.Repo.RepoSpec()
	if _, err := method(ctx, &repoSpec); err != nil {
		return err
	}

	// Always redirect back to the repo main page after enabling.
	http.Redirect(w, r, router.Rel.URLToRepo(rc.Repo.URI).String(), http.StatusSeeOther)
	return nil
}
