package federated

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/ext/github"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

var githubProfiles = conf.GetenvBool("SG_GITHUB_PROFILES")

func CustomUsersGet(ctx context.Context, v1 *sourcegraph.UserSpec, s sourcegraph.UsersServer) (*sourcegraph.User, error) {
	// githubWrap is invoked prior to return. If the input error is nil then the
	// parameters are directly returned. Otherwise, GitHub is queried for the
	// user and the parameters are ignored if $SG_GITHUB_PROFILES=true.
	//
	// Note: This hack is to support Sourcegraph.com's needs (it is not needed
	// generally by all local Sourcegraphs).
	githubWrap := func(u *sourcegraph.User, e error) (*sourcegraph.User, error) {
		// If there is no error, or we're not serving GitHub profiles, so just
		// pass through the values.
		if e == nil || !githubProfiles {
			return u, e
		}
		return new(github.Users).Get(ctx, *v1)
	}

	return githubWrap(s.Get(ctx, v1))
}
