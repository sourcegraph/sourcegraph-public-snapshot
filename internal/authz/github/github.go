package github

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Provider implements authz.Provider for GitHub repository permissions.
type Provider struct {
	urn      string
	client   client
	codeHost *extsvc.CodeHost
	// groupsCache may be nil if group caching is disabled (negative TTL)
	groupsCache *cachedGroups
}

type ProviderOptions struct {
	// If a GitHubClient is not provided, one is constructed from GitHubURL
	GitHubClient *github.V3Client
	GitHubURL    *url.URL

	BaseToken      string
	GroupsCacheTTL time.Duration
}

func NewProvider(urn string, opts ProviderOptions) *Provider {
	if opts.GitHubClient == nil {
		apiURL, _ := github.APIRoot(opts.GitHubURL)
		opts.GitHubClient = github.NewV3Client(apiURL, &auth.OAuthBearerToken{Token: opts.BaseToken}, nil)
	}

	codeHost := extsvc.NewCodeHost(opts.GitHubURL, extsvc.TypeGitHub)
	return &Provider{
		urn:         urn,
		codeHost:    codeHost,
		groupsCache: newGroupPermsCache(urn, codeHost, opts.GroupsCacheTTL),
		client:      &ClientAdapter{V3Client: opts.GitHubClient},
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
	required := p.requiredAuthScopes()
	if len(required) > 0 {
		scopes, err := p.client.GetAuthenticatedUserOAuthScopes(context.Background())
		if err != nil {
			problems = append(problems, fmt.Sprintf("Additional auth scopes are required, but failed to check authenticated scopes: %+v", err))
		}

		gotScopes := make(map[string]struct{})
		for _, value := range scopes {
			gotScopes[value] = struct{}{}
		}
		for _, value := range required {
			if _, found := gotScopes[value.scope]; !found {
				return []string{value.message}
			}
		}
	}
	return problems
}

type requiredScope struct {
	scope   string
	message string
}

func (p *Provider) requiredAuthScopes() []requiredScope {
	scopes := []requiredScope{}

	if p.groupsCache != nil {
		// Needs extra scope to pull group permissions
		scopes = append(scopes, requiredScope{
			scope: "read:org",
			message: "Scope 'read:org' is required to enable `authorization.groupsCacheTTL` - " +
				"please provide a `token` with the required scopes, or try updating the [**site configuration**](/site-admin/configuration)'s " +
				"corresponding entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) to enable `allowGroupsPermissionsSync`.",
		})
	}

	return scopes
}

// fetchUserPermsByToken fetches all the private repo ids that the token can access.
//
// This may return a partial result if an error is encountered, e.g. via rate limits.
func (p *Provider) fetchUserPermsByToken(ctx context.Context, accountID extsvc.AccountID, token string, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	// 🚨 SECURITY: Use user token is required to only list repositories the user has access to.
	client := p.client.WithToken(token)

	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	const repoSetSize = 100
	perms := &authz.ExternalUserPermissions{
		Exacts: make([]extsvc.RepoID, 0, repoSetSize),
	}
	seenRepos := make(map[extsvc.RepoID]struct{}, repoSetSize)

	// addRepoToUserPerms checks if the given repos are already tracked before adding it to perms.
	addRepoToUserPerms := func(repos ...extsvc.RepoID) {
		for _, repo := range repos {
			if _, exists := seenRepos[repo]; !exists {
				seenRepos[repo] = struct{}{}
				perms.Exacts = append(perms.Exacts, repo)
			}
		}
	}

	// If groups caching is enabled, we sync just a subset of direct affiliations - we let
	// other permissions ('organization' affiliation) be sync'd by teams/orgs.
	affiliations := []github.RepositoryAffiliation{github.AffiliationOwner, github.AffiliationCollaborator}
	if p.groupsCache == nil {
		// Otherwise, sync all direct affiliations.
		affiliations = nil
	}

	// Sync direct affiliations
	hasNextPage := true
	var err error
	for page := 1; hasNextPage; page++ {
		var repos []*github.Repository
		repos, hasNextPage, _, err = client.ListAffiliatedRepositories(ctx, github.VisibilityPrivate, page, affiliations...)
		if err != nil {
			return perms, errors.Wrap(err, "list repos for user")
		}

		for _, r := range repos {
			addRepoToUserPerms(extsvc.RepoID(r.ID))
		}
	}

	// If groups caching is disabled, we are done.
	if p.groupsCache == nil {
		return perms, nil
	}

	// Now, we look for groups this user belongs to that give access to additional
	// repositories.
	groups, err := p.getUserAffiliatedGroups(ctx, client, opts)
	if err != nil {
		return perms, errors.Wrap(err, "get groups affiliated with user")
	}

	// Get repos from groups, cached if possible.
	for _, group := range groups {
		// If repositories is empty, perform a full sync
		if len(group.Repositories) > 0 {
			// If a valid cached value was found, use it
			addRepoToUserPerms(group.Repositories...)
			// Add to users iff non-empty
			if len(group.Users) > 0 {
				// If this user's membership in this group is not noted yet, add it
				hasUser := false
				for _, user := range group.Users {
					if user == accountID {
						hasUser = true
						break
					}
				}
				if !hasUser {
					group.Users = append(group.Users, accountID)
					p.groupsCache.setGroup(group)
				}
			}
			continue
		}

		// Initialize group for a fresh sync
		group.Repositories = make([]extsvc.RepoID, 0, repoSetSize)

		// Sync group
		isOrg := group.Team == ""
		hasNextPage = true
		for page := 1; hasNextPage; page++ {
			var repos []*github.Repository
			if isOrg {
				repos, hasNextPage, _, err = p.client.ListOrgRepositories(ctx, group.Org, page, "")
			} else {
				repos, hasNextPage, _, err = p.client.ListTeamRepositories(ctx, group.Org, group.Team, page)
			}
			if err != nil {
				// Add and return what we've found on this page but don't persist group
				// to cache
				for _, r := range repos {
					addRepoToUserPerms(extsvc.RepoID(r.ID))
				}
				return perms, errors.Wrap(err, "list repos for group")
			}
			// Add results to both group (for persistence) and permissions for user
			for _, r := range repos {
				repoID := extsvc.RepoID(r.ID)
				group.Repositories = append(group.Repositories, repoID)
				addRepoToUserPerms(repoID)
			}
		}

		// Persist repos affiliated with group to cache
		p.groupsCache.setGroup(group)
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
func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
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

	return p.fetchUserPermsByToken(ctx, extsvc.AccountID(account.AccountID), tok.AccessToken, opts)
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
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
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
	const userPageSize = 100
	userIDs := make([]extsvc.AccountID, 0, userPageSize)
	seenUsers := make(map[extsvc.AccountID]struct{}, userPageSize)

	// addUserToRepoPerms checks if the given users are already tracked before adding it to perms.
	addUserToRepoPerms := func(users ...extsvc.AccountID) {
		for _, user := range users {
			if _, exists := seenUsers[user]; !exists {
				seenUsers[user] = struct{}{}
				userIDs = append(userIDs, user)
			}
		}
	}

	// If groups caching is enabled, we sync just direct affiliations, and sync org/team
	// collaborators separately from cache
	affiliation := github.AffiliationDirect
	if p.groupsCache == nil {
		// Otherwise, sync all affiliations.
		affiliation = ""
	}

	// Sync collaborators
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var err error
		var users []*github.Collaborator
		users, hasNextPage, err = p.client.ListRepositoryCollaborators(ctx, owner, name, page, affiliation)
		if err != nil {
			return userIDs, err
		}

		for _, u := range users {
			addUserToRepoPerms(extsvc.AccountID(strconv.FormatInt(u.DatabaseID, 10)))
		}
	}

	// If groups caching is disabled, we are done.
	if p.groupsCache == nil {
		return userIDs, nil
	}

	// Get groups affiliated with this repo.
	groups, err := p.getRepoAffiliatedGroups(ctx, owner, name, opts)
	if err != nil {
		return userIDs, errors.Wrap(err, "get groups affiliated with repo")
	}

	// Perform a fresh sync with groups that need a sync.
	repoID := extsvc.RepoID(repo.ID)
	for _, group := range groups {
		// If users is empty, perform a full sync
		if len(group.Users) > 0 {
			// Just use cache if available and not invalidated
			addUserToRepoPerms(group.Users...)
			// Add to repositories iff non-empty
			if len(group.Repositories) > 0 {
				// If this repo's membership in this group is not noted yet, add it
				hasRepo := false
				for _, user := range group.Repositories {
					if user == repoID {
						hasRepo = true
						break
					}
				}
				if !hasRepo {
					group.Repositories = append(group.Repositories, repoID)
					p.groupsCache.setGroup(group.cachedGroup)
				}
			}
		} else {
			// Cache was invalidated or is not yet populated, perform a sync
			hasNextPage := true
			for page := 1; hasNextPage; page++ {
				var members []*github.Collaborator
				if group.Team == "" {
					members, hasNextPage, err = p.client.ListOrganizationMembers(ctx, owner, page, group.adminsOnly)
				} else {
					members, hasNextPage, err = p.client.ListTeamMembers(ctx, owner, group.Team, page)
				}
				if err != nil {
					return append(userIDs, group.Users...), err
				}
				for _, u := range members {
					// Add results to both group (for persistence) and permissions for user
					accountID := extsvc.AccountID(strconv.FormatInt(u.DatabaseID, 10))
					group.Users = append(group.Users, accountID)
					addUserToRepoPerms(accountID)
				}
			}

			// Persist group
			p.groupsCache.setGroup(group.cachedGroup)
		}
	}

	return userIDs, nil
}

// getUserAffiliatedGroups retrieves affiliated organizations and teams for the given client
// with token. Returned groups are populated from cache if a valid value is available.
func (p *Provider) getUserAffiliatedGroups(ctx context.Context, clientWithToken client, opts authz.FetchPermsOptions) ([]cachedGroup, error) {
	groups := make([]cachedGroup, 0)
	seenGroups := make(map[string]struct{})

	// syncGroup adds the given group to the list of groups to cache, pulling values from
	// cache where available.
	syncGroup := func(org, team string) {
		if team != "" {
			// If a team's repos is a subset of an organization's, don't sync. Because when an organization
			// has at least default read permissions, a team's repos will always be a strict subset
			// of the organization's.
			if _, exists := seenGroups[team]; exists {
				return
			}
		}
		cachedPerms, exists := p.groupsCache.getGroup(org, team)
		if exists && opts.InvalidateCaches {
			// invalidate this cache
			p.groupsCache.invalidateGroup(&cachedPerms)
		}
		seenGroups[cachedPerms.key()] = struct{}{}
		groups = append(groups, cachedPerms)
	}
	var err error

	// Get orgs
	hasNextPage := true
	for page := 1; hasNextPage; page++ {
		var orgs []github.OrgDetailsAndMembership
		orgs, hasNextPage, _, err = clientWithToken.GetAuthenticatedUserOrgsDetailsAndMembership(ctx, page)
		if err != nil {
			return groups, err
		}
		for _, org := range orgs {
			// If THIS USER can view this org's repos, we add the entire org to the sync list
			if canViewOrgRepos(&org) {
				syncGroup(org.Login, "")
			}
		}
	}

	// Get teams
	hasNextPage = true
	for page := 1; hasNextPage; page++ {
		var teams []*github.Team
		teams, hasNextPage, _, err = clientWithToken.GetAuthenticatedUserTeams(ctx, page)
		if err != nil {
			return groups, err
		}
		for _, team := range teams {
			// only sync teams with repos
			if team.ReposCount > 0 && team.Organization != nil {
				syncGroup(team.Organization.Login, team.Slug)
			}
		}
	}

	return groups, nil
}

func (p *Provider) getRepoAffiliatedGroups(ctx context.Context, owner, name string, opts authz.FetchPermsOptions) (groups []struct {
	cachedGroup
	adminsOnly bool
}, err error) {
	// Check if repo belongs in an org
	org, err := p.client.GetOrganization(ctx, owner)
	if err != nil {
		if github.IsNotFound(err) {
			// Owner is most likely not an org. User repos don't have teams or org permissions,
			// so we are done - this is fine, so don't propagate error.
			return groups, nil
		}
		return
	}

	// indicate if a group should be sync'd
	syncGroup := func(owner, team string, adminsOnly bool) {
		group, exists := p.groupsCache.getGroup(owner, team)
		if exists && opts.InvalidateCaches {
			// invalidate this cache
			p.groupsCache.invalidateGroup(&group)
		}
		groups = append(groups, struct {
			cachedGroup
			adminsOnly bool
		}{cachedGroup: group, adminsOnly: adminsOnly})
	}

	allOrgMembersCanRead := canViewOrgRepos(&github.OrgDetailsAndMembership{OrgDetails: org})
	if allOrgMembersCanRead {
		// If all members of this org can view this repo, indicate that all members should
		// be sync'd.
		syncGroup(org.Login, "", false)
	} else {
		// Sync only admins of this org
		syncGroup(org.Login, "", true)

		// Also check for teams involved in repo, and indicate all groups should be sync'd.
		hasNextPage := true
		for page := 1; hasNextPage; page++ {
			var teams []*github.Team
			teams, hasNextPage, err = p.client.ListRepositoryTeams(ctx, owner, name, page)
			if err != nil {
				return
			}
			for _, t := range teams {
				syncGroup(org.Login, t.Slug, false)
			}
		}
	}

	return groups, nil
}
