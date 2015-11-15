package federated

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/fed/discover"
	"src.sourcegraph.com/sourcegraph/server/local"
)

type contextKey int

const (
	alreadyLookedUpRepo contextKey = iota
)

// lookupRepo performs discovery on the repo path and returns a new
// Context with the appropriate services and stores to use when
// performing operations on the repo. It also returns info about the
// outcome of the discovery process.
//
// It may modify the repository path (in the `repo` arg) that should
// be used.
//
// If the returned Context is nil, then the caller's underlying
// service should be used.
func lookupRepo(ctx context.Context, repo *string) (context.Context, discover.Info, error) {
	// Set NoSetCacheKey to avoid setting a cache-control trailer for
	// Repos.Get, as opposed to the actual gRPC endpoint
	ctx = context.WithValue(ctx, local.NoSetCacheKey, struct{}{})

	if _, err := local.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: *repo}); errcode.GRPC(err) == codes.NotFound {
		if ctx.Value(alreadyLookedUpRepo) != nil {
			// Avoid infinite cycles.
			return nil, nil, err
		}
		ctx = context.WithValue(ctx, alreadyLookedUpRepo, struct{}{})

		info, err2 := discover.Repo(ctx, *repo)
		if err2 != nil {
			// Return original error from local.Repos.Get unless
			// the discovery error was unexpected.
			if !discover.IsNotFound(err2) {
				err = err2
			}
			return nil, nil, err
		}

		// Chop off hostname portion of repo.
		//
		// TODO(sqs!): doesn't work with single-path-component
		// repos. Make this actually use the new repo origin+path
		// stuff in the Graph Federation doc on the Google Drive.
		//
		// TODO(sqs!): also doesn't work with github repos. hacky.
		if !strings.HasPrefix(*repo, githubcli.Config.Host()) {
			*repo = (*repo)[strings.Index(*repo, "/")+1:]
		}

		ctx, err := info.NewContext(ctx)
		return ctx, info, err
	} else if err != nil {
		return nil, nil, err
	}

	// Fall back to the caller's underlying service.
	return nil, nil, nil
}

// RepoContext wraps lookupRepo and discards the discover.Info return
// value. It is used in the codegenned federated method
// implementations (where the Info is not needed).
func RepoContext(ctx context.Context, repo *string) (context.Context, error) {
	ctx, _, err := lookupRepo(ctx, repo)
	return ctx, err
}

func lookupUser(ctx context.Context, user sourcegraph.UserSpec) (context.Context, discover.Info, error) {
	if authutil.ActiveFlags.IsLocal() || authutil.ActiveFlags.IsLDAP() {
		return nil, nil, nil
	}
	if user.Domain == "" {
		if !fed.Config.IsRoot {
			user.Domain = fed.Config.RootURL().Host
		} else {
			return nil, nil, nil
		}
	}
	info, err := discover.SiteURL(ctx, user.Domain)
	if err != nil {
		return nil, nil, err
	}
	ctx, err = info.NewContext(ctx)
	return ctx, info, err
}

// UserContext wraps lookupUser and discards the discover.Info return
// value. It is used in the codegenned federated method
// implementations (where the Info is not needed).
func UserContext(ctx context.Context, user sourcegraph.UserSpec) (context.Context, error) {
	ctx, _, err := lookupUser(ctx, user)
	return ctx, err
}
