package github

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Provider implements authz.Provider for GitHub repository permissions.
type Provider struct {
	urn         string
	client      client
	groupsCache *groupsCache
	codeHost    *extsvc.CodeHost
}

func NewProvider(urn string, githubURL *url.URL, baseToken string, client *github.V3Client) *Provider {
	if client == nil {
		apiURL, _ := github.APIRoot(githubURL)
		client = github.NewV3Client(apiURL, &auth.OAuthBearerToken{Token: baseToken}, nil)
	}
	codeHost := extsvc.NewCodeHost(githubURL, extsvc.TypeGitHub)

	return &Provider{
		urn:         urn,
		codeHost:    codeHost,
		groupsCache: newGroupPermsCache(urn, codeHost),
		client:      &ClientAdapter{V3Client: client},
	}
}

var _ authz.Provider = (*Provider)(nil)

// FetchAccount implements the authz.Provider interface. It always returns nil, because the GitHub
// API doesn't currently provide a way to fetch user by external SSO account.
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

func (p *Provider) Validate() (problems []string) {
	return nil
}

// FetchUserPermsByToken fetches all the private repo ids that the token can access.
//
// This may return a partial result if an error is encountered, e.g. via rate limits.
func (p *Provider) FetchUserPermsByToken(ctx context.Context, token string) (*authz.ExternalUserPermissions, error) {
	// 🚨 SECURITY: Use user token is required to only list repositories the user has access to.
	client := p.client.WithToken(token)

	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	const repoSetSize = 100
	perms := &authz.ExternalUserPermissions{
		Exacts: make([]extsvc.RepoID, 0, repoSetSize),
	}
	seenRepos := make(map[extsvc.RepoID]bool, repoSetSize)

	// appendRepo checks if the given repos are already tracked before adding it to perms.
	appendRepo := func(repos ...extsvc.RepoID) {
		for _, repo := range repos {
			if _, exists := seenRepos[repo]; !exists {
				perms.Exacts = append(perms.Exacts, repo)
			}
		}
	}

	// First, we list repositories a user has direct access to as a owner or direct collaborator.
	// We let other permissions ('organization' affiliation) be sync'd by teams/orgs.
	hasNextPage := true
	var err error
	for page := 1; hasNextPage; page++ {
		var repos []*github.Repository
		repos, hasNextPage, _, err = client.ListAffiliatedRepositories(ctx, github.VisibilityPrivate, page,
			github.AffiliationOwner, github.AffiliationCollaborator)
		if err != nil {
			return perms, err
		}

		for _, r := range repos {
			appendRepo(extsvc.RepoID(r.ID))
		}
	}

	// Now, we check for groups this user belongs to that give access to additional
	// repositories.
	//
	// Track groups in this map of "org/team":group so that we don't check GitHub teams
	// that this user already has access to through Organization membership.
	groups := make(map[string]cachedGroup)
	// Get orgs
	hasNextPage = true
	for page := 1; hasNextPage; page++ {
		var orgs []*github.OrgDetails
		orgs, hasNextPage, _, err = client.GetAuthenticatedUserOrgsDetails(ctx, page)
		if err != nil {
			return perms, err
		}
		for _, org := range orgs {
			if org.DefaultRepositoryPermission != "none" {
				groups[org.Login] = cachedGroup{
					Org: org.Login,
				}
			}
		}
	}
	// Get teams
	hasNextPage = true
	for page := 1; hasNextPage; page++ {
		var teams []*github.Team
		teams, hasNextPage, _, err = client.GetAuthenticatedUserTeams(ctx, page)
		if err != nil {
			return perms, err
		}
		for _, team := range teams {
			// if a team's repos is a subset of a organization's, don't sync
			if _, exists := groups[team.Organization.Login]; !exists {
				if team.ReposCount > 0 {
					groups[fmt.Sprintf("%s/%s", team.Organization.Login, team.Slug)] = cachedGroup{
						Org:  team.Organization.Login,
						Team: team.Slug,
					}
				}
			}
		}
	}

	// Get repos from groups, cached if possible.
	for _, group := range groups {
		groupPerms, exists := p.groupsCache.getGroupPermsFromCache(group.Org, group.Team)
		if exists {
			appendRepo(groupPerms.Repositories...)
			continue
		}

		// TODO: fetch repos, append to perms.Exacts, push to cache
	}

	return perms, nil
}

// FetchUserPerms returns a list of repository IDs (on code host) that the given account
// has read access on the code host. The repository ID has the same value as it would be
// used as api.ExternalRepoSpec.ID. The returned list only includes private repository IDs.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://developer.github.com/v3/repos/#list-repositories-for-the-authenticated-user
func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account) (*authz.ExternalUserPermissions, error) {
	if account == nil {
		return nil, errors.New("no account provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, account) {
		return nil, errors.Errorf("not a code host of the account: want %q but have %q",
			account.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	_, tok, err := github.GetExternalAccountData(&account.AccountData)
	if err != nil {
		return nil, errors.Wrap(err, "get external account data")
	} else if tok == nil {
		return nil, errors.New("no token found in the external account data")
	}

	return p.FetchUserPermsByToken(ctx, tok.AccessToken)
}

// FetchRepoPerms returns a list of user IDs (on code host) who have read access to
// the given project on the code host. The user ID has the same value as it would
// be used as extsvc.Account.AccountID. The returned list includes both direct access
// and inherited from the organization membership.
//
// This method may return partial but valid results in case of error, and it is up to
// callers to decide whether to discard.
//
// API docs: https://developer.github.com/v4/object/repositorycollaboratorconnection/
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository) ([]extsvc.AccountID, error) {
	if repo == nil {
		return nil, errors.New("no repository provided")
	} else if !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternalRepoSpec) {
		return nil, errors.Errorf("not a code host of the repository: want %q but have %q",
			repo.ServiceID, p.codeHost.ServiceID)
	}

	// NOTE: We do not store port or scheme in our URI, so stripping the hostname alone is enough.
	nameWithOwner := strings.TrimPrefix(repo.URI, p.codeHost.BaseURL.Hostname())
	nameWithOwner = strings.TrimPrefix(nameWithOwner, "/")

	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return nil, errors.Wrap(err, "split nameWithOwner")
	}

	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	userIDs := make([]extsvc.AccountID, 0, 100)
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var err error
		var users []*github.Collaborator
		users, hasNextPage, err = p.client.ListRepositoryCollaborators(ctx, owner, name, page)
		if err != nil {
			return userIDs, err
		}

		for _, u := range users {
			userIDs = append(userIDs, extsvc.AccountID(strconv.FormatInt(u.DatabaseID, 10)))
		}
	}

	return userIDs, nil
}
