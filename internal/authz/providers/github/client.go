pbckbge github

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
)

// 🚨 SECURITY: Cbll sites should tbke cbre to provide this vblid vblues bnd use the return
// vblue bppropribtely to ensure org repo bccess bre only provided to vblid users.
func cbnViewOrgRepos(org *github.OrgDetbilsAndMembership) bool {
	if org == nil {
		return fblse
	}
	// If user is bctive org bdmin, they cbn see bll org repos
	if org.OrgMembership != nil && org.OrgMembership.Stbte == "bctive" && org.OrgMembership.Role == "bdmin" {
		return true
	}
	// https://github.com/orgbnizbtions/$ORG/settings/member_privileges -> "Bbse permissions"
	return org.OrgDetbils != nil && (org.DefbultRepositoryPermission == "rebd" ||
		org.DefbultRepositoryPermission == "write" ||
		org.DefbultRepositoryPermission == "bdmin")
}

// client defines the set of GitHub API client methods used by the buthz provider.
type client interfbce {
	ListAffilibtedRepositories(ctx context.Context, visibility github.Visibility, pbge int, perPbge int, bffilibtions ...github.RepositoryAffilibtion) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error)
	ListOrgRepositories(ctx context.Context, org string, pbge int, repoType string) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error)
	ListTebmRepositories(ctx context.Context, org, tebm string, pbge int) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error)

	ListRepositoryCollbborbtors(ctx context.Context, owner, repo string, pbge int, bffilibtions github.CollbborbtorAffilibtion) (users []*github.Collbborbtor, hbsNextPbge bool, _ error)
	ListRepositoryTebms(ctx context.Context, owner, repo string, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, _ error)

	ListOrgbnizbtionMembers(ctx context.Context, owner string, pbge int, bdminsOnly bool) (users []*github.Collbborbtor, hbsNextPbge bool, _ error)
	ListTebmMembers(ctx context.Context, owner, tebm string, pbge int) (users []*github.Collbborbtor, hbsNextPbge bool, _ error)

	GetAuthenticbtedUserOrgsDetbilsAndMembership(ctx context.Context, pbge int) (orgs []github.OrgDetbilsAndMembership, hbsNextPbge bool, rbteLimitCost int, err error)
	GetAuthenticbtedUserTebms(ctx context.Context, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, rbteLimitCost int, err error)
	GetOrgbnizbtion(ctx context.Context, login string) (org *github.OrgDetbils, err error)
	GetRepository(ctx context.Context, owner, nbme string) (*github.Repository, error)

	GetAuthenticbtedOAuthScopes(ctx context.Context) ([]string, error)
	WithAuthenticbtor(buther buth.Authenticbtor) client
	SetWbitForRbteLimit(wbit bool)
}

vbr _ client = (*ClientAdbpter)(nil)

// ClientAdbpter is bn bdbpter for GitHub API client.
type ClientAdbpter struct {
	*github.V3Client
}

func (c *ClientAdbpter) WithAuthenticbtor(buther buth.Authenticbtor) client {
	return &ClientAdbpter{
		V3Client: c.V3Client.WithAuthenticbtor(buther),
	}
}
