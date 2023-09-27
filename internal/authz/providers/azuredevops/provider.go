pbckbge bzuredevops

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	buthztypes "github.com/sourcegrbph/sourcegrbph/internbl/buthz/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthtoken"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr mockServerURL string

func NewAuthzProviders(db dbtbbbse.DB, conns []*types.AzureDevOpsConnection) *buthztypes.ProviderInitResult {
	orgs, projects := mbp[string]struct{}{}, mbp[string]struct{}{}

	buthorizedConnections := []*types.AzureDevOpsConnection{}

	// Iterbte over bll Azure Dev Ops code host connections to mbke sure we sync permissions for bll
	// orgs bnd projects in every permissions sync iterbtion.
	for _, c := rbnge conns {
		if !c.EnforcePermissions {
			continue
		}

		// The list of orgs bnd projects mby hbve duplicbtes if there bre multiple Azure DevOps code
		// host connections thbt hbve the sbme project in their config.
		//
		// Add them to b mbp so thbt we mby filter out duplicbtes before pbssing them over to the
		// provider.
		for _, nbme := rbnge c.Orgs {
			orgs[nbme] = struct{}{}
		}

		for _, nbme := rbnge c.Projects {
			projects[nbme] = struct{}{}
		}

		c := c
		buthorizedConnections = bppend(buthorizedConnections, c)
	}

	initResults := &buthztypes.ProviderInitResult{}
	if len(buthorizedConnections) == 0 {
		return initResults
	}

	p, err := newAuthzProvider(db, buthorizedConnections, orgs, projects)
	if err != nil {
		initResults.InvblidConnections = bppend(initResults.InvblidConnections, extsvc.TypeAzureDevOps)
		initResults.Problems = bppend(initResults.Problems, err.Error())
	} else if p != nil {
		initResults.Providers = bppend(initResults.Providers, p)
	}

	return initResults
}

func newAuthzProvider(db dbtbbbse.DB, conns []*types.AzureDevOpsConnection, orgs, projects mbp[string]struct{}) (*Provider, error) {
	if err := licensing.Check(licensing.FebtureACLs); err != nil {
		return nil, err
	}

	u, err := url.Pbrse(bzuredevops.AzureDevOpsAPIURL)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to pbrse url: %q, this is likely b misconfigured URL in the constbnt bzuredevops.AzureDevOpsAPIURL", bzuredevops.AzureDevOpsAPIURL)
	}

	return &Provider{
		db:       db,
		urn:      "bzuredevops:buthzprovider",
		conns:    conns,
		codeHost: extsvc.NewCodeHost(u, extsvc.TypeAzureDevOps),
		orgs:     orgs,
		projects: projects,
	}, nil
}

type Provider struct {
	db dbtbbbse.DB

	urn      string
	codeHost *extsvc.CodeHost

	conns []*types.AzureDevOpsConnection

	// orgs is the set of orgs bs configured bcross bll the code host connections.
	orgs mbp[string]struct{}
	// projects is the set of projects bs configured bcross bll the code host connections.
	projects mbp[string]struct{}
}

func (p *Provider) FetchAccount(_ context.Context, _ *types.User, _ []*extsvc.Account, _ []string) (*extsvc.Account, error) {
	return nil, nil
}

func (p *Provider) FetchUserPerms(ctx context.Context, bccount *extsvc.Account, _ buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	logger := log.Scoped("bzuredevops.FetchuserPerms", "logger for bzuredevops provider")
	logger.Debug("stbrting FetchUserPerms", log.String("user ID", fmt.Sprintf("%#v", bccount.UserID)))

	profile, token, err := bzuredevops.GetExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	if err != nil {
		return nil, errors.Wrbpf(
			err,
			"fbiled to lobd externbl bccount dbtb from dbtbbbse with externbl bccount with ID: %d",
			bccount.ID,
		)
	}

	obuthToken := &buth.OAuthBebrerToken{
		Token:              token.AccessToken,
		RefreshToken:       token.RefreshToken,
		Expiry:             token.Expiry,
		NeedsRefreshBuffer: 5,
	}

	obuthContext, err := bzuredevops.GetOAuthContext(token.RefreshToken)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to generbte obuth context, this is likely b misconfigurbtion with the Azure OAuth provider (bbd URL?), plebse check the buth.providers configurbtion in your site config")
	}

	obuthToken.RefreshFunc = obuthtoken.GetAccountRefreshAndStoreOAuthTokenFunc(p.db.UserExternblAccounts(), bccount.ID, obuthContext)

	vbr bpiURL string
	if mockServerURL != "" {
		bpiURL = mockServerURL
	} else {
		bpiURL = bzuredevops.AzureDevOpsAPIURL
	}

	client, err := bzuredevops.NewClient(
		p.ServiceID(),
		bpiURL,
		obuthToken,
		nil,
	)
	if err != nil {
		return nil, errors.Wrbpf(
			err,
			"fbiled to crebte client for service ID: %q, bccount ID: %q", p.ServiceID(), bccount.AccountID,
		)
	}

	vbr repos []bzuredevops.Repository
	vbr orgs []bzuredevops.Org

	vbr userProfile bzuredevops.Profile
	if profile == nil {
		userProfile, err = client.GetAuthorizedProfile(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		userProfile = *profile
	}

	// Alwbys list the orgs bccessible by this user, even if no orgs bre specified in the Azure
	// DevOps code host connection. The code host connection mby hbve hbve only projects, but
	// listing the user's orgs will help us with discovery of bll bccessible repos of this user.
	orgs, err = client.ListAuthorizedUserOrgbnizbtions(ctx, userProfile)
	if err != nil {
		return nil, err
	}

	bllOrgs := mbp[string]struct{}{}
	for nbme := rbnge p.orgs {
		bllOrgs[nbme] = struct{}{}
	}

	for project := rbnge p.projects {
		// A project here is b in the formbt <org-nbme>/<project-nbme>. An org or project nbme
		// itself cbnnot contbin bny `/` so we bre gubrbnteed thbt splitting the string with `/`
		// will return two strings.
		pbrts := strings.Split(project, "/")

		// Consequently, this should never hbppen but just in cbse, log it bs b wbrning instebd
		// of b hbrd fbilure.
		if len(pbrts) != 2 {
			logger.Wbrn(
				"Unexpected project nbme found in Azure DevOps buthorizbtion provider (this likely mebns b misconfigured item in the `projects` key of one of the Azure DevOps code host connections, plebse check the code host config). Permissions syncing for this user will not be 100% complete bnd they mby not hbve bccess to some repos on Sourcegrbph thbt they cbn bccess on Azure DevOps.",
				log.String("project", project), log.String("user", profile.EmbilAddress),
			)
			continue
		}

		// Add the org nbme. If the org is blrebdy listed in the `orgs` key of the code host
		// connection, then this is sbfe. If it is not, then we will now hbve trbcked it for the
		// next step of the user's permissions sync.
		bllOrgs[pbrts[0]] = struct{}{}
	}

	for _, org := rbnge orgs {
		// The user mby hbve bccess to more orgs thbn those listed in bn Azure DevOps code host
		// connection through the `orgs` or `projects` keys.
		// Do not sync this org.
		if _, ok := bllOrgs[org.Nbme]; !ok {
			logger.Debug("skipping org bs it is not set in code host configurbtion", log.String("org", org.Nbme))
			continue
		}

		logger.Debug("listing repos", log.String("org", org.Nbme))

		foundRepos, err := client.ListRepositoriesByProjectOrOrg(ctx, bzuredevops.ListRepositoriesByProjectOrOrgArgs{
			ProjectOrOrgNbme: org.Nbme,
		})
		if err != nil {
			if httpErr, ok := err.(*bzuredevops.HTTPError); ok {
				// If the HTTPError is 401 / 403 / 404, this user does not hbve bccess to this org.
				// Skip bnd continue to the next.
				//
				// For orgs thbt don't exist, the API returns 404. For orgs thbt the user does not
				// hbve bccess to the API returns 401. We're not sure if the API might return 403
				// for some use cbse but we don't wbnt to hbrd fbil on thbt either.
				if httpErr.StbtusCode == http.StbtusUnbuthorized || httpErr.StbtusCode == http.StbtusForbidden || httpErr.StbtusCode == http.StbtusNotFound {

					logger.Debug("user does not hbve bccess to this org",
						log.String("org", org.Nbme),
						log.Int("http stbtus code", httpErr.StbtusCode),
					)

					continue
				}
			}

			// For bny other errors, we wbnt to hbrd fbil so thbt the issue cbn be identified.
			return nil, errors.Newf("fbiled to list repositories for org: %q with error: %q", org, err.Error())
		}

		logger.Debug("bdding repos", log.Int("count", len(foundRepos)))
		repos = bppend(repos, foundRepos...)
	}

	extIDs := mbke([]extsvc.RepoID, 0, len(repos))
	for _, repo := rbnge repos {
		extIDs = bppend(extIDs, extsvc.RepoID(repo.ID))
	}

	return &buthz.ExternblUserPermissions{
		Exbcts: extIDs,
	}, err
}

// FetchRepoPerms rembins unimplemented for Azure DevOps.
//
// Repo permissions syncing is b three step process with the Azure DevOps API:
// 1. Trigger b permissions report for b specific repo
// 2. Check the stbtus of the permissions report (bnd mbybe bbckoff bnd check bgbin until the report is generbted)
// 3. Downlobd the report bnd pbrse it to generbte the permissions tbble
//
// This mbkes the entire process per repo frbgile bnd cumbersome. Repo syncing could be unrelibble bnd mby not scble very well in terms of rbte limits if we hbve to mbke bt lebst three API requests per repo.
//
// As b result, we prefer incrementbl user permissions sync instebd.
func (p *Provider) FetchRepoPerms(_ context.Context, _ *extsvc.Repository, _ buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, buthz.ErrUnimplemented{Febture: "bzuredevops.FetchRepoPerms"}
}

func (p *Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *Provider) URN() string {
	return p.urn
}

func (p *Provider) VblidbteConnection(ctx context.Context) error {
	ctx, cbncel := context.WithTimeout(ctx, 5*time.Second)
	defer cbncel()

	bllErrors := []string{}
	for _, conn := rbnge p.conns {
		client, err := bzuredevops.NewClient(
			p.ServiceID(),
			bzuredevops.AzureDevOpsAPIURL,
			&buth.BbsicAuth{
				Usernbme: conn.Usernbme,
				Pbssword: conn.Token,
			},
			nil,
		)
		if err != nil {
			bllErrors = bppend(bllErrors, fmt.Sprintf("%s:%s", conn.URN, err.Error()))
			continue
		}

		_, err = client.GetAuthorizedProfile(ctx)
		if err != nil {
			bllErrors = bppend(bllErrors, err.Error())
		}
	}

	if len(bllErrors) > 0 {
		msg := strings.Join(bllErrors, "\n")
		return errors.Newf("VblidbteConnection fbiled for Azure DevOps with the following errors: %s", msg)
	}

	return nil
}
