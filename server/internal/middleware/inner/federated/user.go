package federated

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/ext/github"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
)

var githubProfiles = conf.GetenvBool("SG_GITHUB_PROFILES")

func CustomUsersGet(ctx context.Context, v1 *sourcegraph.UserSpec, s sourcegraph.UsersServer) (*sourcegraph.User, error) {
	tmp := *v1
	tmp.Domain = ""

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
		return new(github.Users).Get(ctx, tmp)
	}

	if authutil.ActiveFlags.IsLocal() {
		return githubWrap(s.Get(ctx, v1))
	}

	if authutil.ActiveFlags.IsLDAP() {
		return s.Get(ctx, v1)
	}

	if !fed.Config.IsRoot && v1.Domain == "" {
		v1.Domain = fed.Config.RootURL().Host
	}

	ctx2, err := UserContext(ctx, *v1)
	if err != nil {
		return nil, err
	}
	if ctx2 == nil {
		return s.Get(ctx, v1)
	}
	ctx = ctx2

	user, err := svc.Users(ctx).Get(ctx, &tmp)
	if user != nil {
		user.Domain = v1.Domain
	}
	return user, err
}
