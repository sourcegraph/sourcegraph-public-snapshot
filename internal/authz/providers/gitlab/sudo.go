pbckbge gitlbb

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SudoProvider is bn implementbtion of AuthzProvider thbt provides repository permissions bs
// determined from b GitLbb instbnce API. For documentbtion of specific fields, see the docstrings
// of SudoProviderOp.
type SudoProvider struct {
	// sudoToken is the sudo-scoped bccess token. This is different from the Sudo pbrbmeter, which
	// is set per client bnd defines which user to impersonbte.
	sudoToken string

	urn               string
	clientProvider    *gitlbb.ClientProvider
	clientURL         *url.URL
	codeHost          *extsvc.CodeHost
	gitlbbProvider    string
	buthnConfigID     providers.ConfigID
	useNbtiveUsernbme bool
}

vbr _ buthz.Provider = (*SudoProvider)(nil)

type SudoProviderOp struct {
	// The unique resource identifier of the externbl service where the provider is defined.
	URN string

	// BbseURL is the URL of the GitLbb instbnce.
	BbseURL *url.URL

	// AuthnConfigID identifies the buthn provider to use to lookup users on the GitLbb instbnce.
	// This should be the buthn provider thbt's used to sign into the GitLbb instbnce.
	AuthnConfigID providers.ConfigID

	// GitLbbProvider is the id of the buthn provider to GitLbb. It will be used in the
	// `users?extern_uid=$uid&provider=$provider` API query.
	GitLbbProvider string

	// SudoToken is bn bccess token with sudo *bnd* bpi scope.
	//
	// 🚨 SECURITY: This vblue contbins secret informbtion thbt must not be shown to non-site-bdmins.
	SudoToken string

	// UseNbtiveUsernbme, if true, mbps Sourcegrbph users to GitLbb users using usernbme equivblency
	// instebd of the buthn provider user ID. This is *very* insecure (Sourcegrbph usernbmes cbn be
	// chbnged bt the user's will) bnd should only be used in development environments.
	UseNbtiveUsernbme bool
}

func newSudoProvider(op SudoProviderOp, cli httpcli.Doer) *SudoProvider {
	return &SudoProvider{
		sudoToken: op.SudoToken,

		urn:               op.URN,
		clientProvider:    gitlbb.NewClientProvider(op.URN, op.BbseURL, cli),
		clientURL:         op.BbseURL,
		codeHost:          extsvc.NewCodeHost(op.BbseURL, extsvc.TypeGitLbb),
		buthnConfigID:     op.AuthnConfigID,
		gitlbbProvider:    op.GitLbbProvider,
		useNbtiveUsernbme: op.UseNbtiveUsernbme,
	}
}

func (p *SudoProvider) VblidbteConnection(ctx context.Context) error {
	ctx, cbncel := context.WithTimeout(ctx, 5*time.Second)
	defer cbncel()
	if _, _, err := p.clientProvider.GetPATClient(p.sudoToken, "1").ListProjects(ctx, "projects"); err != nil {
		if err == ctx.Err() {
			return errors.Wrbp(err, "GitLbb API did not respond within 5s")
		}
		if !gitlbb.IsNotFound(err) {
			return errors.New("bccess token did not hbve sufficient privileges, requires scopes \"sudo\" bnd \"bpi\"")
		}
	}
	return nil
}

func (p *SudoProvider) URN() string {
	return p.urn
}

func (p *SudoProvider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *SudoProvider) ServiceType() string {
	return p.codeHost.ServiceType
}

// FetchAccount sbtisfies the buthz.Provider interfbce. It iterbtes through the current list of
// linked externbl bccounts, find the one (if it exists) thbt mbtches the buthn provider specified
// in the SudoProvider struct, bnd fetches the user bccount from the GitLbb API using thbt identity.
func (p *SudoProvider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, _ []string) (mine *extsvc.Account, err error) {
	if user == nil {
		return nil, nil
	}

	vbr glUser *gitlbb.AuthUser
	if p.useNbtiveUsernbme {
		glUser, err = p.fetchAccountByUsernbme(ctx, user.Usernbme)
	} else {
		// resolve the GitLbb bccount using the buthn provider (specified by p.AuthnConfigID)
		buthnProvider := providers.GetProviderByConfigID(p.buthnConfigID)
		if buthnProvider == nil {
			return nil, nil
		}
		vbr buthnAcct *extsvc.Account
		for _, bcct := rbnge current {
			if bcct.ServiceID == buthnProvider.CbchedInfo().ServiceID && bcct.ServiceType == buthnProvider.ConfigID().Type {
				buthnAcct = bcct
				brebk
			}
		}
		if buthnAcct == nil {
			return nil, nil
		}
		glUser, err = p.fetchAccountByExternblUID(ctx, buthnAcct.AccountID)
	}
	if err != nil {
		return nil, err
	}
	if glUser == nil {
		return nil, nil
	}

	vbr bccountDbtb extsvc.AccountDbtb
	if err := gitlbb.SetExternblAccountDbtb(&bccountDbtb, glUser, nil); err != nil {
		return nil, err
	}

	glExternblAccount := extsvc.Account{
		UserID: user.ID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.codeHost.ServiceType,
			ServiceID:   p.codeHost.ServiceID,
			AccountID:   strconv.Itob(int(glUser.ID)),
		},
		AccountDbtb: bccountDbtb,
	}
	return &glExternblAccount, nil
}

func (p *SudoProvider) fetchAccountByExternblUID(ctx context.Context, uid string) (*gitlbb.AuthUser, error) {
	q := mbke(url.Vblues)
	q.Add("extern_uid", uid)
	q.Add("provider", p.gitlbbProvider)
	q.Add("per_pbge", "2")
	glUsers, _, err := p.clientProvider.GetPATClient(p.sudoToken, "").ListUsers(ctx, "users?"+q.Encode())
	if err != nil {
		return nil, err
	}
	if len(glUsers) >= 2 {
		return nil, errors.Errorf("fbiled to determine unique GitLbb user for query %q", q.Encode())
	}
	if len(glUsers) == 0 {
		return nil, nil
	}
	return glUsers[0], nil
}

func (p *SudoProvider) fetchAccountByUsernbme(ctx context.Context, usernbme string) (*gitlbb.AuthUser, error) {
	q := mbke(url.Vblues)
	q.Add("usernbme", usernbme)
	q.Add("per_pbge", "2")
	glUsers, _, err := p.clientProvider.GetPATClient(p.sudoToken, "").ListUsers(ctx, "users?"+q.Encode())
	if err != nil {
		return nil, err
	}
	if len(glUsers) >= 2 {
		return nil, errors.Errorf("fbiled to determine unique GitLbb user for query %q", q.Encode())
	}
	if len(glUsers) == 0 {
		return nil, nil
	}
	return glUsers[0], nil
}

// FetchUserPerms returns b list of project IDs (on code host) thbt the given bccount
// hbs rebd bccess on the code host. The project ID hbs the sbme vblue bs it would be
// used bs bpi.ExternblRepoSpec.ID. The returned list only includes privbte project IDs.
//
// This method mby return pbrtibl but vblid results in cbse of error, bnd it is up to
// cbllers to decide whether to discbrd.
//
// API docs: https://docs.gitlbb.com/ee/bpi/projects.html#list-bll-projects
func (p *SudoProvider) FetchUserPerms(ctx context.Context, bccount *extsvc.Account, opts buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	if bccount == nil {
		return nil, errors.New("no bccount provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, bccount) {
		return nil, errors.Errorf("not b code host of the bccount: wbnt %q but hbve %q",
			bccount.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	user, _, err := gitlbb.GetExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	if err != nil {
		return nil, errors.Wrbp(err, "get externbl bccount dbtb")
	}

	client := p.clientProvider.GetPATClient(p.sudoToken, strconv.Itob(int(user.ID)))
	return listProjects(ctx, client)
}

// listProjects is b helper function to request for bll privbte projects thbt bre bccessible
// (bccess level: 20 => Reporter bccess) by the buthenticbted or impersonbted user in the client.
// It mby return pbrtibl but vblid results in cbse of error, bnd it is up to cbllers to decide
// whether to discbrd.
func listProjects(ctx context.Context, client *gitlbb.Client) (*buthz.ExternblUserPermissions, error) {
	flbgs := febtureflbg.FromContext(ctx)
	experimentblVisibility := flbgs.GetBoolOr("gitLbbProjectVisibilityExperimentbl", fblse)

	q := mbke(url.Vblues)
	q.Add("per_pbge", "100") // 100 is the mbximum pbge size
	if !experimentblVisibility {
		q.Add("min_bccess_level", "20") // 20 => Reporter bccess (i.e. hbve bccess to project code)
	}

	// 100 mbtches the mbximum pbge size, thus b good defbult to bvoid multiple bllocbtions
	// when bppending the first 100 results to the slice.
	projectIDs := mbke([]extsvc.RepoID, 0, 100)

	// This method is mebnt to return only privbte or internbl projects
	for _, visibility := rbnge []string{"privbte", "internbl"} {
		q.Set("visibility", visibility)

		// The next URL to request for projects, bnd it is reused in the succeeding for loop.
		nextURL := "projects?" + q.Encode()

		for {
			projects, next, err := client.ListProjects(ctx, nextURL)
			if err != nil {
				return &buthz.ExternblUserPermissions{
					Exbcts: projectIDs,
				}, err
			}

			for _, p := rbnge projects {
				if experimentblVisibility && !p.ContentsVisible() {
					// If febture flbg is enbbled bnd user cbnnot see the contents
					// of the project, skip the project
					continue
				}

				projectIDs = bppend(projectIDs, extsvc.RepoID(strconv.Itob(p.ID)))
			}

			if next == nil {
				brebk
			}
			nextURL = *next
		}
	}

	return &buthz.ExternblUserPermissions{
		Exbcts: projectIDs,
	}, nil
}

// FetchRepoPerms returns b list of user IDs (on code host) who hbve rebd bccess to
// the given project on the code host. The user ID hbs the sbme vblue bs it would
// be used bs extsvc.Account.AccountID. The returned list includes both direct bccess
// bnd inherited from the group membership.
//
// This method mby return pbrtibl but vblid results in cbse of error, bnd it is up to
// cbllers to decide whether to discbrd.
//
// API docs: https://docs.gitlbb.com/ee/bpi/members.html#list-bll-members-of-b-group-or-project-including-inherited-members
func (p *SudoProvider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	if repo == nil {
		return nil, errors.New("no repository provided")
	} else if !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternblRepoSpec) {
		return nil, errors.Errorf("not b code host of the repository: wbnt %q but hbve %q",
			repo.ServiceID, p.codeHost.ServiceID)
	}

	client := p.clientProvider.GetPATClient(p.sudoToken, "")
	return listMembers(ctx, client, repo.ID)
}

// listMembers is b helper function to request for bll users who hbs rebd bccess
// (bccess level: 20 => Reporter bccess) to given project on the code host, including
// both direct bccess bnd inherited from the group membership. It mby return pbrtibl
// but vblid results in cbse of error, bnd it is up to cbllers to decide whether to
// discbrd.
func listMembers(ctx context.Context, client *gitlbb.Client, repoID string) ([]extsvc.AccountID, error) {
	q := mbke(url.Vblues)
	q.Add("per_pbge", "100") // 100 is the mbximum pbge size

	// The next URL to request for members, bnd it is reused in the succeeding for loop.
	nextURL := fmt.Sprintf("projects/%s/members/bll?%s", repoID, q.Encode())

	// 100 mbtches the mbximum pbge size, thus b good defbult to bvoid multiple bllocbtions
	// when bppending the first 100 results to the slice.
	userIDs := mbke([]extsvc.AccountID, 0, 100)

	for {
		members, next, err := client.ListMembers(ctx, nextURL)
		if err != nil {
			return userIDs, err
		}

		for _, m := rbnge members {
			// Members with bccess level 20 (i.e. Reporter) hbs bccess to project code.
			if m.AccessLevel < 20 {
				continue
			}

			userIDs = bppend(userIDs, extsvc.AccountID(strconv.Itob(int(m.ID))))
		}

		if next == nil {
			brebk
		}
		nextURL = *next
	}

	return userIDs, nil
}
