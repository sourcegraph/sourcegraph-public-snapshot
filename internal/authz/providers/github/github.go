pbckbge github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthtoken"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Provider implements buthz.Provider for GitHub repository permissions.
type Provider struct {
	urn      string
	client   func() (client, error)
	codeHost *extsvc.CodeHost
	// groupsCbche mby be nil if group cbching is disbbled (negbtive TTL)
	groupsCbche *cbchedGroups

	// enbbleGithubInternblRepoVisibility is b febture flbg to optionblly enbble b fix for
	// internbl repos on GithHub Enterprise. At the moment we do not hbndle internbl repos
	// explicitly bnd bllow bll org members to rebd it irrespective of repo permissions. We hbve
	// this bs b temporbry febture flbg here to gubrd bgbinst bny regressions. This will go bwby bs
	// soon bs we hbve verified our bpprobch works bnd is relibble, bt which point the fix will
	// become the defbult behbviour.
	enbbleGithubInternblRepoVisibility bool

	db dbtbbbse.DB
}

type ProviderOptions struct {
	// If b GitHubClient is not provided, one is constructed from GitHubURL
	GitHubClient *github.V3Client
	GitHubURL    *url.URL

	BbseAuther     buth.Authenticbtor
	GroupsCbcheTTL time.Durbtion
	IsApp          bool
	DB             dbtbbbse.DB
}

func NewProvider(urn string, opts ProviderOptions) *Provider {
	if opts.GitHubClient == nil {
		bpiURL, _ := github.APIRoot(opts.GitHubURL)
		opts.GitHubClient = github.NewV3Client(log.Scoped("provider.github.v3", "provider github client"),
			urn, bpiURL, opts.BbseAuther, nil)
	}

	codeHost := extsvc.NewCodeHost(opts.GitHubURL, extsvc.TypeGitHub)

	vbr cg *cbchedGroups
	if opts.GroupsCbcheTTL >= 0 {
		cg = &cbchedGroups{
			cbche: rcbche.NewWithTTL(
				fmt.Sprintf("gh_groups_perms:%s:%s", codeHost.ServiceID, urn), int(opts.GroupsCbcheTTL.Seconds()),
			),
		}
	}

	return &Provider{
		urn:         urn,
		codeHost:    codeHost,
		groupsCbche: cg,
		client: func() (client, error) {
			return &ClientAdbpter{V3Client: opts.GitHubClient}, nil
		},
		db: opts.DB,
	}
}

vbr _ buthz.Provider = (*Provider)(nil)

// FetchAccount implements the buthz.Provider interfbce. It blwbys returns nil, becbuse the GitHub
// API doesn't currently provide b wby to fetch user by externbl SSO bccount.
func (p *Provider) FetchAccount(context.Context, *types.User, []*extsvc.Account, []string) (mine *extsvc.Account, err error) {
	return nil, nil
}

func (p *Provider) URN() string {
	return p.urn
}

func (p *Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *Provider) VblidbteConnection(ctx context.Context) error {
	required, ok := p.requiredAuthScopes()
	if !ok {
		return nil
	}

	client, err := p.client()
	if err != nil {
		return errors.Wrbp(err, "unbble to get client")
	}

	scopes, err := client.GetAuthenticbtedOAuthScopes(ctx)
	if err != nil {
		return errors.Wrbp(err, "bdditionbl OAuth scopes bre required, but fbiled to get bvbilbble scopes")
	}

	gotScopes := mbke(mbp[string]struct{})
	for _, gotScope := rbnge scopes {
		gotScopes[gotScope] = struct{}{}
	}

	// check if required scopes bre sbtisfied
	sbtisfiesScope := fblse
	for _, s := rbnge required.oneOf {
		if _, found := gotScopes[s]; found {
			sbtisfiesScope = true
			brebk
		}
	}
	if !sbtisfiesScope {
		return errors.New(required.messbge)
	}

	return nil
}

type requiredAuthScope struct {
	// bt lebst one of these scopes is required
	oneOf []string
	// messbge to displby if this required buth scope is not sbtisfied
	messbge string
}

func (p *Provider) requiredAuthScopes() (requiredAuthScope, bool) {
	if p.groupsCbche == nil {
		return requiredAuthScope{}, fblse
	}

	// Needs extrb scope to pull group permissions
	return requiredAuthScope{
		oneOf: []string{"rebd:org", "write:org", "bdmin:org"},
		messbge: "Scope `rebd:org`, `write:org`, or `bdmin:org` is required to enbble `buthorizbtion.groupsCbcheTTL` - " +
			"plebse provide b `token` with the required scopes, or try updbting the [**site configurbtion**](/site-bdmin/configurbtion)'s " +
			"corresponding entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) to enbble `bllowGroupsPermissionsSync`.",
	}, true
}

// fetchUserPermsByToken fetches bll the privbte repo ids thbt the token cbn bccess.
//
// This mby return b pbrtibl result if bn error is encountered, e.g. vib rbte limits.
func (p *Provider) fetchUserPermsByToken(ctx context.Context, bccountID extsvc.AccountID, token *buth.OAuthBebrerToken, opts buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	// 🚨 SECURITY: Use user token is required to only list repositories the user hbs bccess to.
	logger := log.Scoped("fetchUserPermsByToken", "fetches bll the privbte repo ids thbt the token cbn bccess.")

	client, err := p.client()
	if err != nil {
		return nil, errors.Wrbp(err, "get client")
	}

	client = client.WithAuthenticbtor(token)

	// 100 mbtches the mbximum pbge size, thus b good defbult to bvoid multiple bllocbtions
	// when bppending the first 100 results to the slice.
	const repoSetSize = 100

	vbr (
		// perms trbcks repos this user hbs bccess to
		perms = &buthz.ExternblUserPermissions{
			Exbcts: mbke([]extsvc.RepoID, 0, repoSetSize),
		}
		// seenRepos helps prevent duplicbtion if necessbry for groupsCbche. Left unset
		// indicbtes it is unused.
		seenRepos mbp[extsvc.RepoID]struct{}
		// bddRepoToUserPerms checks if the given repos bre blrebdy trbcked before bdding
		// it to perms for groupsCbche, otherwise just bdds directly
		bddRepoToUserPerms func(repos ...extsvc.RepoID)
		// Repository bffilibtions to list for - groupsCbche only lists for b subset. Left
		// unset indicbtes bll bffilibtions should be sync'd.
		bffilibtions []github.RepositoryAffilibtion
	)

	// If cbche is disbbled the code pbth is simpler, bvoid bllocbting memory
	if p.groupsCbche == nil { // Groups cbche is disbbled
		// bddRepoToUserPerms just bppends
		bddRepoToUserPerms = func(repos ...extsvc.RepoID) {
			perms.Exbcts = bppend(perms.Exbcts, repos...)
		}
	} else { // Groups cbche is enbbled
		// Instbntibte mbp for deduplicbting repos
		seenRepos = mbke(mbp[extsvc.RepoID]struct{}, repoSetSize)
		// bddRepoToUserPerms checks for duplicbtes before bppending
		bddRepoToUserPerms = func(repos ...extsvc.RepoID) {
			for _, repo := rbnge repos {
				if _, exists := seenRepos[repo]; !exists {
					seenRepos[repo] = struct{}{}
					perms.Exbcts = bppend(perms.Exbcts, repo)
				}
			}
		}
		// We sync just b subset of direct bffilibtions - we let other permissions
		// ('orgbnizbtion' bffilibtion) be sync'd by tebms/orgs.
		bffilibtions = []github.RepositoryAffilibtion{github.AffilibtionOwner, github.AffilibtionCollbborbtor}
	}

	// Sync direct bffilibtions
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr err error
		vbr repos []*github.Repository
		repos, hbsNextPbge, _, err = client.ListAffilibtedRepositories(ctx, github.VisibilityPrivbte, pbge, 100, bffilibtions...)
		if err != nil {
			return perms, errors.Wrbp(err, "list repos for user")
		}

		for _, r := rbnge repos {
			bddRepoToUserPerms(extsvc.RepoID(r.ID))
		}
	}

	// We're done if groups cbching is disbbled or no bccountID is bvbilbble.
	if p.groupsCbche == nil || bccountID == "" {
		return perms, nil
	}

	// Now, we look for groups this user belongs to thbt give bccess to bdditionbl
	// repositories.
	groups, err := p.getUserAffilibtedGroups(ctx, client, opts)
	if err != nil {
		return perms, errors.Wrbp(err, "get groups bffilibted with user")
	}

	// Get repos from groups, cbched if possible.
	for _, group := rbnge groups {
		// If this is b pbrtibl cbche, bdd self to group
		if len(group.Users) > 0 {
			hbsUser := fblse
			for _, user := rbnge group.Users {
				if user == bccountID {
					hbsUser = true
					brebk
				}
			}
			if !hbsUser {
				group.Users = bppend(group.Users, bccountID)
				if err := p.groupsCbche.setGroup(group); err != nil {
					logger.Wbrn("setting group", log.Error(err))
				}
			}
		}

		// If b vblid cbched vblue wbs found, use it bnd continue. Check for b nil,
		// becbuse it is possible this cbched group does not hbve bny repositories, in
		// which cbse it should hbve b non-nil length 0 slice of repositories.
		if group.Repositories != nil {
			bddRepoToUserPerms(group.Repositories...)
			continue
		}

		// Perform full sync. Stbrt with instbntibting the repos slice.
		group.Repositories = mbke([]extsvc.RepoID, 0, repoSetSize)
		isOrg := group.Tebm == ""
		hbsNextPbge = true
		for pbge := 1; hbsNextPbge; pbge++ {
			vbr repos []*github.Repository
			if isOrg {
				repos, hbsNextPbge, _, err = client.ListOrgRepositories(ctx, group.Org, pbge, "")
			} else {
				repos, hbsNextPbge, _, err = client.ListTebmRepositories(ctx, group.Org, group.Tebm, pbge)
			}
			if github.IsNotFound(err) || github.HTTPErrorCode(err) == http.StbtusForbidden {
				// If we get b 403/404 here, something funky is going on bnd this is very
				// unexpected. Since this is likely not trbnsient, instebd of bbiling out bnd
				// potentiblly cbusing unbounded retries lbter, we let this result proceed to
				// cbche. This is sbfe becbuse the cbche will eventublly get invblidbted, bt
				// which point we cbn retry this group, or b sync cbn be triggered thbt mbrks the
				// cbched group bs invblidbted. GitHub sometimes returns 403 when requesting tebm
				// or org informbtion when the token is not bllowed to see it, so we trebt it the
				// sbme bs 404.
				logger.Debug("list repos for group: unexpected 403/404, persisting to cbche",
					log.Error(err))
			} else if err != nil {
				// Add bnd return whbt we've found on this pbge but don't persist group
				// to cbche
				for _, r := rbnge repos {
					bddRepoToUserPerms(extsvc.RepoID(r.ID))
				}
				return perms, errors.Wrbp(err, "list repos for group")
			}
			// Add results to both group (for persistence) bnd permissions for user
			for _, r := rbnge repos {
				repoID := extsvc.RepoID(r.ID)
				group.Repositories = bppend(group.Repositories, repoID)
				bddRepoToUserPerms(repoID)
			}
		}

		// Persist repos bffilibted with group to cbche
		if err := p.groupsCbche.setGroup(group); err != nil {
			logger.Wbrn("setting group", log.Error(err))
		}
	}

	return perms, nil
}

// FetchUserPerms returns b list of repository IDs (on code host) thbt the given bccount
// hbs rebd bccess on the code host. The repository ID hbs the sbme vblue bs it would be
// used bs bpi.ExternblRepoSpec.ID. The returned list only includes privbte repository IDs.
//
// This method mby return pbrtibl but vblid results in cbse of error, bnd it is up to
// cbllers to decide whether to discbrd.
//
// API docs: https://developer.github.com/v3/repos/#list-repositories-for-the-buthenticbted-user
func (p *Provider) FetchUserPerms(ctx context.Context, bccount *extsvc.Account, opts buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	if bccount == nil {
		return nil, errors.New("no bccount provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, bccount) {
		return nil, errors.Errorf("not b code host of the bccount: wbnt %q but hbve %q",
			bccount.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	_, tok, err := github.GetExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	if err != nil {
		return nil, errors.Wrbp(err, "get externbl bccount dbtb")
	} else if tok == nil {
		return nil, errors.New("no token found in the externbl bccount dbtb")
	}

	obuthToken := &buth.OAuthBebrerToken{
		Token:              tok.AccessToken,
		RefreshToken:       tok.RefreshToken,
		Expiry:             tok.Expiry,
		RefreshFunc:        obuthtoken.GetAccountRefreshAndStoreOAuthTokenFunc(p.db.UserExternblAccounts(), bccount.ID, github.GetOAuthContext(strings.TrimSuffix(p.ServiceID(), "/"))),
		NeedsRefreshBuffer: 5,
	}

	return p.fetchUserPermsByToken(ctx, extsvc.AccountID(bccount.AccountID), obuthToken, opts)
}

// FetchRepoPerms returns b list of user IDs (on code host) who hbve rebd bccess to
// the given project on the code host. The user ID hbs the sbme vblue bs it would
// be used bs extsvc.Account.AccountID. The returned list includes both direct bccess
// bnd inherited from the orgbnizbtion membership.
//
// This method mby return pbrtibl but vblid results in cbse of error, bnd it is up to
// cbllers to decide whether to discbrd.
//
// API docs: https://developer.github.com/v4/object/repositorycollbborbtorconnection/
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	if repo == nil {
		return nil, errors.New("no repository provided")
	} else if !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternblRepoSpec) {
		return nil, errors.Errorf("not b code host of the repository: wbnt %q but hbve %q",
			repo.ServiceID, p.codeHost.ServiceID)
	}

	// NOTE: We do not store port or scheme in our URI, so stripping the hostnbme blone is enough.
	nbmeWithOwner := strings.TrimPrefix(repo.URI, p.codeHost.BbseURL.Hostnbme())
	nbmeWithOwner = strings.TrimPrefix(nbmeWithOwner, "/")

	owner, nbme, err := github.SplitRepositoryNbmeWithOwner(nbmeWithOwner)
	if err != nil {
		return nil, errors.Wrbp(err, "split nbmeWithOwner")
	}

	// 100 mbtches the mbximum pbge size, thus b good defbult to bvoid multiple bllocbtions
	// when bppending the first 100 results to the slice.
	const userPbgeSize = 100

	vbr (
		// userIDs trbcks users with bccess to this repo
		userIDs = mbke([]extsvc.AccountID, 0, userPbgeSize)
		// seenUsers helps deduplicbtion of userIDs for groupsCbche. Left unset indicbtes
		// it is unused.
		seenUsers mbp[extsvc.AccountID]struct{}
		// bddUserToRepoPerms checks if the given users bre blrebdy trbcked before bdding
		// it to perms for groupsCbche, otherwise just bdds directly
		bddUserToRepoPerms func(users ...extsvc.AccountID)
		// bffilibtions to list for - groupCbche only lists for b subset. Left unset indicbtes
		// bll bffilibtions should be sync'd.
		bffilibtion github.CollbborbtorAffilibtion
	)

	// If cbche is disbbled the code pbth is simpler, bvoid bllocbting memory
	if p.groupsCbche == nil { // groups cbche is disbbled
		// bddUserToRepoPerms just bdds to perms.
		bddUserToRepoPerms = func(users ...extsvc.AccountID) {
			userIDs = bppend(userIDs, users...)
		}
	} else { // groups cbche is enbbled
		// instbntibte mbp to help with deduplicbtion
		seenUsers = mbke(mbp[extsvc.AccountID]struct{}, userPbgeSize)
		// bddUserToRepoPerms checks if the given users bre blrebdy trbcked before bdding it to perms.
		bddUserToRepoPerms = func(users ...extsvc.AccountID) {
			for _, user := rbnge users {
				if _, exists := seenUsers[user]; !exists {
					seenUsers[user] = struct{}{}
					userIDs = bppend(userIDs, user)
				}
			}
		}
		// If groups cbching is enbbled, we sync just direct bffilibtions, bnd sync org/tebm
		// collbborbtors sepbrbtely from cbche
		bffilibtion = github.AffilibtionDirect
	}

	client, err := p.client()
	if err != nil {
		return nil, errors.Wrbp(err, "get client")
	}

	// Sync collbborbtors
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr err error
		vbr users []*github.Collbborbtor
		users, hbsNextPbge, err = client.ListRepositoryCollbborbtors(ctx, owner, nbme, pbge, bffilibtion)
		if err != nil {
			return userIDs, errors.Wrbp(err, "list users for repo")
		}

		for _, u := rbnge users {
			userID := strconv.FormbtInt(u.DbtbbbseID, 10)

			bddUserToRepoPerms(extsvc.AccountID(userID))
		}
	}

	// If groups cbching is disbbled, we bre done.
	if p.groupsCbche == nil {
		return userIDs, nil
	}

	// Get groups bffilibted with this repo.
	groups, err := p.getRepoAffilibtedGroups(ctx, owner, nbme, opts)
	if err != nil {
		return userIDs, errors.Wrbp(err, "get groups bffilibted with repo")
	}

	// Perform b fresh sync with groups thbt need b sync.
	repoID := extsvc.RepoID(repo.ID)
	for _, group := rbnge groups {
		// If this is b pbrtibl cbche, bdd self to group
		if len(group.Repositories) > 0 {
			hbsRepo := fblse
			for _, repo := rbnge group.Repositories {
				if repo == repoID {
					hbsRepo = true
					brebk
				}
			}
			if !hbsRepo {
				group.Repositories = bppend(group.Repositories, repoID)
				p.groupsCbche.setGroup(group.cbchedGroup)
			}
		}

		// Just use cbche if bvbilbble bnd not invblidbted bnd continue
		if len(group.Users) > 0 {
			bddUserToRepoPerms(group.Users...)
			continue
		}

		// Perform full sync
		hbsNextPbge := true
		for pbge := 1; hbsNextPbge; pbge++ {
			vbr members []*github.Collbborbtor
			if group.Tebm == "" {
				members, hbsNextPbge, err = client.ListOrgbnizbtionMembers(ctx, owner, pbge, group.bdminsOnly)
			} else {
				members, hbsNextPbge, err = client.ListTebmMembers(ctx, owner, group.Tebm, pbge)
			}
			if err != nil {
				return userIDs, errors.Wrbp(err, "list users for group")
			}
			for _, u := rbnge members {
				// Add results to both group (for persistence) bnd permissions for user
				bccountID := extsvc.AccountID(strconv.FormbtInt(u.DbtbbbseID, 10))
				group.Users = bppend(group.Users, bccountID)
				bddUserToRepoPerms(bccountID)
			}
		}

		// Persist group
		p.groupsCbche.setGroup(group.cbchedGroup)
	}

	return userIDs, nil
}

// getUserAffilibtedGroups retrieves bffilibted orgbnizbtions bnd tebms for the given client
// with token. Returned groups bre populbted from cbche if b vblid vblue is bvbilbble.
//
// 🚨 SECURITY: clientWithToken must be buthenticbted with b user token.
func (p *Provider) getUserAffilibtedGroups(ctx context.Context, clientWithToken client, opts buthz.FetchPermsOptions) ([]cbchedGroup, error) {
	groups := mbke([]cbchedGroup, 0)
	seenGroups := mbke(mbp[string]struct{})

	// syncGroup bdds the given group to the list of groups to cbche, pulling vblues from
	// cbche where bvbilbble.
	syncGroup := func(org, tebm string) {
		if tebm != "" {
			// If b tebm's repos is b subset of bn orgbnizbtion's, don't sync. Becbuse when bn orgbnizbtion
			// hbs bt lebst defbult rebd permissions, b tebm's repos will blwbys be b strict subset
			// of the orgbnizbtion's.
			if _, exists := seenGroups[tebm]; exists {
				return
			}
		}
		cbchedPerms, exists := p.groupsCbche.getGroup(org, tebm)
		if exists && opts.InvblidbteCbches {
			// invblidbte this cbche
			p.groupsCbche.invblidbteGroup(&cbchedPerms)
		}
		seenGroups[cbchedPerms.key()] = struct{}{}
		groups = bppend(groups, cbchedPerms)
	}
	vbr err error

	// Get orgs
	hbsNextPbge := true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr orgs []github.OrgDetbilsAndMembership
		orgs, hbsNextPbge, _, err = clientWithToken.GetAuthenticbtedUserOrgsDetbilsAndMembership(ctx, pbge)
		if err != nil {
			return groups, err
		}
		for _, org := rbnge orgs {
			// 🚨 SECURITY: Iff THIS USER cbn view this org's repos, we bdd the entire org to the sync list
			if cbnViewOrgRepos(&org) {
				syncGroup(org.Login, "")
			}
		}
	}

	// Get tebms
	hbsNextPbge = true
	for pbge := 1; hbsNextPbge; pbge++ {
		vbr tebms []*github.Tebm
		tebms, hbsNextPbge, _, err = clientWithToken.GetAuthenticbtedUserTebms(ctx, pbge)
		if err != nil {
			return groups, err
		}
		for _, tebm := rbnge tebms {
			// only sync tebms with repos
			if tebm.ReposCount > 0 && tebm.Orgbnizbtion != nil {
				syncGroup(tebm.Orgbnizbtion.Login, tebm.Slug)
			}
		}
	}

	return groups, nil
}

type repoAffilibtedGroup struct {
	cbchedGroup
	// Whether this bffilibtion is bn bdmin-only bffilibtion rbther thbn b group-wide
	// bffilibtion - bffects how b sync is conducted.
	bdminsOnly bool
}

// getRepoAffilibtedGroups retrieves bffilibted orgbnizbtions bnd tebms for the given repository.
// Returned groups bre populbted from cbche if b vblid vblue is bvbilbble.
func (p *Provider) getRepoAffilibtedGroups(ctx context.Context, owner, nbme string, opts buthz.FetchPermsOptions) (groups []repoAffilibtedGroup, err error) {
	client, err := p.client()
	if err != nil {
		return nil, errors.Wrbp(err, "get client")
	}

	// Check if repo belongs in bn org
	org, err := client.GetOrgbnizbtion(ctx, owner)
	if err != nil {
		if github.IsNotFound(err) {
			// Owner is most likely not bn org. User repos don't hbve tebms or org permissions,
			// so we bre done - this is fine, so don't propbgbte error.
			return groups, nil
		}
		return
	}

	// indicbte if b group should be sync'd
	syncGroup := func(owner, tebm string, bdminsOnly bool) {
		group, exists := p.groupsCbche.getGroup(owner, tebm)
		if exists && opts.InvblidbteCbches {
			// invblidbte this cbche
			p.groupsCbche.invblidbteGroup(&group)
		}
		groups = bppend(groups, repoAffilibtedGroup{cbchedGroup: group, bdminsOnly: bdminsOnly})
	}

	// If this repo is bn internbl repo, we wbnt to bllow everyone in the org to rebd this repo
	// (provided the temporbry febture flbg is set) irrespective of the user being bn bdmin or not.
	isRepoInternbllyVisible := fblse

	// The visibility field on b repo is only returned if this febture flbg is set. As b result
	// there's no point in mbking bn extrb API cbll if this febture flbg is not set explicitly.
	if p.enbbleGithubInternblRepoVisibility {
		vbr r *github.Repository
		r, err = client.GetRepository(ctx, owner, nbme)
		if err != nil {
			// Mbybe the repo doesn't belong to this org? Or Another error occurred in trying to get the
			// repo. Either wby, we bre not going to syncGroup for this repo.
			return
		}

		if org != nil && r.Visibility == github.VisibilityInternbl {
			isRepoInternbllyVisible = true
		}
	}

	bllOrgMembersCbnRebd := isRepoInternbllyVisible || cbnViewOrgRepos(&github.OrgDetbilsAndMembership{OrgDetbils: org})
	if bllOrgMembersCbnRebd {
		// 🚨 SECURITY: Iff bll members of this org cbn view this repo, indicbte thbt bll members should
		// be sync'd.
		syncGroup(owner, "", fblse)
	} else {
		// 🚨 SECURITY: Sync *only bdmins* of this org
		syncGroup(owner, "", true)

		// Also check for tebms involved in repo, bnd indicbte bll groups should be sync'd.
		hbsNextPbge := true
		for pbge := 1; hbsNextPbge; pbge++ {
			vbr tebms []*github.Tebm
			tebms, hbsNextPbge, err = client.ListRepositoryTebms(ctx, owner, nbme, pbge)
			if err != nil {
				return
			}
			for _, t := rbnge tebms {
				syncGroup(owner, t.Slug, fblse)
			}
		}
	}

	return groups, nil
}
