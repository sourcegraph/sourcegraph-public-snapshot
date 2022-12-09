package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	gh "github.com/google/go-github/v48/github"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Provider implements authz.Provider for GitHub repository permissions.
type Provider struct {
	urn      string
	client   func() (client, error)
	codeHost *extsvc.CodeHost
	// groupsCache may be nil if group caching is disabled (negative TTL)
	groupsCache *cachedGroups

	// enableGithubInternalRepoVisibility is a feature flag to optionally enable a fix for
	// internal repos on GithHub Enterprise. At the moment we do not handle internal repos
	// explicitly and allow all org members to read it irrespective of repo permissions. We have
	// this as a temporary feature flag here to guard against any regressions. This will go away as
	// soon as we have verified our approach works and is reliable, at which point the fix will
	// become the default behaviour.
	enableGithubInternalRepoVisibility bool

	InstallationID *int64

	db database.DB

	baseHTTPClient *http.Client
	baseToken      string
	baseURL        *url.URL
}

type ProviderOptions struct {
	// If a GitHubClient is not provided, one is constructed from GitHubURL
	GitHubClient *github.V3Client
	GitHubURL    *url.URL

	BaseToken      string
	GroupsCacheTTL time.Duration
	IsApp          bool
	DB             database.DB
	BaseHTTPClient *http.Client
}

func NewProvider(urn string, opts ProviderOptions) *Provider {
	apiURL, isGitHubDotCom := github.APIRoot(opts.GitHubURL)
	if opts.GitHubClient == nil {
		opts.GitHubClient = github.NewV3Client(log.Scoped("provider.github.v3", "provider github client"),
			urn, apiURL, &auth.OAuthBearerToken{Token: opts.BaseToken}, nil)
	}

	codeHost := extsvc.NewCodeHost(opts.GitHubURL, extsvc.TypeGitHub)

	var cg *cachedGroups
	if opts.GroupsCacheTTL >= 0 {
		cg = &cachedGroups{
			cache: rcache.NewWithTTL(
				fmt.Sprintf("gh_groups_perms:%s:%s", codeHost.ServiceID, urn), int(opts.GroupsCacheTTL.Seconds()),
			),
		}
	}

	var baseURL *url.URL
	if !isGitHubDotCom {
		baseURL = apiURL
	}

	return &Provider{
		urn:         urn,
		codeHost:    codeHost,
		groupsCache: cg,
		client: func() (client, error) {
			return &ClientAdapter{V3Client: opts.GitHubClient}, nil
		},
		db:             opts.DB,
		baseHTTPClient: opts.BaseHTTPClient,
		baseToken:      opts.BaseToken,
		baseURL:        baseURL,
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

func (p *Provider) ValidateConnection(ctx context.Context) []string {
	required := p.requiredAuthScopes()
	if len(required) == 0 {
		return []string{}
	}

	client, err := p.client()
	if err != nil {
		return []string{
			fmt.Sprintf("Unable to get client: %v", err),
		}
	}

	scopes, err := client.GetAuthenticatedOAuthScopes(ctx)
	if err != nil {
		return []string{
			fmt.Sprintf("Additional OAuth scopes are required, but failed to get available scopes: %+v", err),
		}
	}

	gotScopes := make(map[string]struct{})
	for _, gotScope := range scopes {
		gotScopes[gotScope] = struct{}{}
	}

	var problems []string
	// check if required scopes are satisfied
	for _, requiredScope := range required {
		satisfiesScope := false
		for _, s := range requiredScope.oneOf {
			if _, found := gotScopes[s]; found {
				satisfiesScope = true
				break
			}
		}
		if !satisfiesScope {
			problems = append(problems, requiredScope.message)
		}
	}

	return problems
}

type requiredAuthScope struct {
	// at least one of these scopes is required
	oneOf []string
	// message to display if this required auth scope is not satisfied
	message string
}

func (p *Provider) requiredAuthScopes() []requiredAuthScope {
	scopes := []requiredAuthScope{}

	if p.groupsCache != nil {
		// Needs extra scope to pull group permissions
		scopes = append(scopes, requiredAuthScope{
			oneOf: []string{"read:org", "write:org", "admin:org"},
			message: "Scope `read:org`, `write:org`, or `admin:org` is required to enable `authorization.groupsCacheTTL` - " +
				"please provide a `token` with the required scopes, or try updating the [**site configuration**](/site-admin/configuration)'s " +
				"corresponding entry in [`auth.providers`](https://docs.sourcegraph.com/admin/auth) to enable `allowGroupsPermissionsSync`.",
		})
	}

	return scopes
}

// fetchUserPermsByToken fetches all the private repo ids that the token can access.
//
// This may return a partial result if an error is encountered, e.g. via rate limits.
func (p *Provider) fetchUserPermsByToken(ctx context.Context, cli *gh.Client, accountID extsvc.AccountID, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	// 100 matches the maximum page size, thus a good default to avoid multiple allocations
	// when appending the first 100 results to the slice.
	const repoSetSize = 100

	var (
		// perms tracks repos this user has access to
		perms = &authz.ExternalUserPermissions{
			Exacts: make([]extsvc.RepoID, 0, repoSetSize),
		}
		// seenRepos helps prevent duplication if necessary for groupsCache. Left unset
		// indicates it is unused.
		seenRepos map[extsvc.RepoID]struct{}
		// addRepoToUserPerms checks if the given repos are already tracked before adding
		// it to perms for groupsCache, otherwise just adds directly
		addRepoToUserPerms func(repos ...extsvc.RepoID)
		// Repository affiliations to list for - groupsCache only lists for a subset. Left
		// unset indicates all affiliations should be sync'd.
		affiliations []string
	)

	// If cache is disabled the code path is simpler, avoid allocating memory
	if p.groupsCache == nil { // Groups cache is disabled
		// addRepoToUserPerms just appends
		addRepoToUserPerms = func(repos ...extsvc.RepoID) {
			perms.Exacts = append(perms.Exacts, repos...)
		}
	} else { // Groups cache is enabled
		// Instantiate map for deduplicating repos
		seenRepos = make(map[extsvc.RepoID]struct{}, repoSetSize)
		// addRepoToUserPerms checks for duplicates before appending
		addRepoToUserPerms = func(repos ...extsvc.RepoID) {
			for _, repo := range repos {
				if _, exists := seenRepos[repo]; !exists {
					seenRepos[repo] = struct{}{}
					perms.Exacts = append(perms.Exacts, repo)
				}
			}
		}
		// We sync just a subset of direct affiliations - we let other permissions
		// ('organization' affiliation) be sync'd by teams/orgs.
		affiliations = []string{string(github.AffiliationOwner), string(github.AffiliationCollaborator)}
	}

	// Sync direct affiliations
	listOpts := &gh.RepositoryListOptions{
		Visibility:  "private",
		Affiliation: strings.Join(affiliations, ","),
		Sort:        "created",
		ListOptions: gh.ListOptions{
			Page:    1,
			PerPage: repoSetSize,
		},
	}
	for {
		ghRepos, resp, err := cli.Repositories.List(ctx, "", listOpts)
		if err != nil {
			return perms, errors.Wrap(err, "list repos for user")
		}

		for _, r := range ghRepos {
			addRepoToUserPerms(extsvc.RepoID(r.GetNodeID()))
		}
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	// We're done if groups caching is disabled or no accountID is available.
	if p.groupsCache == nil || accountID == "" {
		return perms, nil
	}

	// Now, we look for groups this user belongs to that give access to additional
	// repositories.
	groups, err := p.getUserAffiliatedGroups(ctx, cli, opts)
	if err != nil {
		return perms, errors.Wrap(err, "get groups affiliated with user")
	}

	logger := log.Scoped("fetchUserPermsByToken", "fetches all the private repo ids that the token can access.")

	// Get repos from groups, cached if possible.
	for _, group := range groups {
		// If this is a partial cache, add self to group
		if len(group.Users) > 0 {
			hasUser := false
			for _, user := range group.Users {
				if user == accountID {
					hasUser = true
					break
				}
			}
			if !hasUser {
				group.Users = append(group.Users, accountID)
				if err := p.groupsCache.setGroup(group); err != nil {
					logger.Warn("setting group", log.Error(err))
				}
			}
		}

		// If a valid cached value was found, use it and continue. Check for a nil,
		// because it is possible this cached group does not have any repositories, in
		// which case it should have a non-nil length 0 slice of repositories.
		if group.Repositories != nil {
			addRepoToUserPerms(group.Repositories...)
			continue
		}

		// Perform full sync. Start with instantiating the repos slice.
		group.Repositories = make([]extsvc.RepoID, 0, repoSetSize)
		isOrg := group.Team == ""
		listOpts := &gh.ListOptions{
			Page:    1,
			PerPage: 100,
		}
		for {
			var repos []*gh.Repository
			var resp *gh.Response
			if isOrg {
				repos, resp, err = cli.Repositories.ListByOrg(ctx, group.Org, &gh.RepositoryListByOrgOptions{Sort: "created", ListOptions: *listOpts})
			} else {
				repos, resp, err = cli.Teams.ListTeamReposBySlug(ctx, group.Org, group.Team, listOpts)
			}
			if github.IsNotFound(err) ||
				github.HTTPErrorCode(err) == http.StatusForbidden ||
				resp != nil && (resp.StatusCode == http.StatusForbidden ||
					resp.StatusCode == http.StatusNotFound) {
				// If we get a 403/404 here, something funky is going on and this is very
				// unexpected. Since this is likely not transient, instead of bailing out and
				// potentially causing unbounded retries later, we let this result proceed to
				// cache. This is safe because the cache will eventually get invalidated, at
				// which point we can retry this group, or a sync can be triggered that marks the
				// cached group as invalidated. GitHub sometimes returns 403 when requesting team
				// or org information when the token is not allowed to see it, so we treat it the
				// same as 404.
				logger.Debug("list repos for group: unexpected 403/404, persisting to cache",
					log.Error(err))
			} else if err != nil {
				// Add and return what we've found on this page but don't persist group
				// to cache
				for _, r := range repos {
					addRepoToUserPerms(extsvc.RepoID(r.GetNodeID()))
				}
				return perms, errors.Wrap(err, "list repos for group")
			}
			// Add results to both group (for persistence) and permissions for user
			for _, r := range repos {
				repoID := extsvc.RepoID(r.GetNodeID())
				group.Repositories = append(group.Repositories, repoID)
				addRepoToUserPerms(repoID)
			}

			if resp.NextPage == 0 {
				break
			}
			listOpts.Page = resp.NextPage
		}

		// Persist repos affiliated with group to cache
		if err := p.groupsCache.setGroup(group); err != nil {
			logger.Warn("setting group", log.Error(err))
		}
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

	_, tok, err := github.GetExternalAccountData(ctx, &account.AccountData)
	if err != nil {
		return nil, errors.Wrap(err, "get external account data")
	} else if tok == nil {
		return nil, errors.New("no token found in the external account data")
	}

	oauthToken := &auth.OAuthBearerToken{
		Token:        tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
	}

	if p.InstallationID != nil && p.db != nil {
		// Only used if created by newAppProvider
		oauthToken.RefreshFunc = database.GetAccountRefreshAndStoreOAuthTokenFunc(p.db, account.ID, github.GetOAuthContext(p.codeHost.BaseURL.String()))
		oauthToken.NeedsRefreshBuffer = 5
	}

	cli, err := github.NewGitHubClientForUserExternalAccount(ctx, account, p.baseHTTPClient, p.baseURL)
	if err != nil {
		return nil, err
	}

	return p.fetchUserPermsByToken(ctx, cli, extsvc.AccountID(account.AccountID), opts)
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
	cli, err := github.NewGitHubClientWithToken(ctx, p.baseToken, p.baseHTTPClient, p.baseURL)
	if err != nil {
		return nil, err
	}

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

	var (
		// userIDs tracks users with access to this repo
		userIDs = make([]extsvc.AccountID, 0, userPageSize)
		// seenUsers helps deduplication of userIDs for groupsCache. Left unset indicates
		// it is unused.
		seenUsers map[extsvc.AccountID]struct{}
		// addUserToRepoPerms checks if the given users are already tracked before adding
		// it to perms for groupsCache, otherwise just adds directly
		addUserToRepoPerms func(users ...extsvc.AccountID)
		// affiliations to list for - groupCache only lists for a subset. Left unset indicates
		// all affiliations should be sync'd.
		affiliation github.CollaboratorAffiliation
	)

	// If cache is disabled the code path is simpler, avoid allocating memory
	if p.groupsCache == nil { // groups cache is disabled
		// addUserToRepoPerms just adds to perms.
		addUserToRepoPerms = func(users ...extsvc.AccountID) {
			userIDs = append(userIDs, users...)
		}
	} else { // groups cache is enabled
		// instantiate map to help with deduplication
		seenUsers = make(map[extsvc.AccountID]struct{}, userPageSize)
		// addUserToRepoPerms checks if the given users are already tracked before adding it to perms.
		addUserToRepoPerms = func(users ...extsvc.AccountID) {
			for _, user := range users {
				if _, exists := seenUsers[user]; !exists {
					seenUsers[user] = struct{}{}
					userIDs = append(userIDs, user)
				}
			}
		}
		// If groups caching is enabled, we sync just direct affiliations, and sync org/team
		// collaborators separately from cache
		affiliation = github.AffiliationDirect
	}

	// Sync collaborators
	listOpts := &gh.ListCollaboratorsOptions{
		Affiliation: string(affiliation),
		ListOptions: gh.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}
	for {
		users, resp, err := cli.Repositories.ListCollaborators(ctx, owner, name, listOpts)
		if err != nil {
			return userIDs, errors.Wrap(err, "list users for repo")
		}

		for _, u := range users {
			userID := strconv.FormatInt(u.GetID(), 10)
			if p.InstallationID != nil {
				userID = strconv.FormatInt(*p.InstallationID, 10) + "/" + userID
			}

			addUserToRepoPerms(extsvc.AccountID(userID))
		}
		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	// If groups caching is disabled, we are done.
	if p.groupsCache == nil {
		return userIDs, nil
	}

	// Get groups affiliated with this repo.
	groups, err := p.getRepoAffiliatedGroups(ctx, owner, name, cli, opts)
	if err != nil {
		return userIDs, errors.Wrap(err, "get groups affiliated with repo")
	}

	// Perform a fresh sync with groups that need a sync.
	repoID := extsvc.RepoID(repo.ID)
	for _, group := range groups {
		// If this is a partial cache, add self to group
		if len(group.Repositories) > 0 {
			hasRepo := false
			for _, repo := range group.Repositories {
				if repo == repoID {
					hasRepo = true
					break
				}
			}
			if !hasRepo {
				group.Repositories = append(group.Repositories, repoID)
				p.groupsCache.setGroup(group.cachedGroup)
			}
		}

		// Just use cache if available and not invalidated and continue
		if len(group.Users) > 0 {
			addUserToRepoPerms(group.Users...)
			continue
		}

		// Perform full sync
		listOpts := gh.ListOptions{
			Page:    1,
			PerPage: 100,
		}
		for {
			var members []*gh.User
			var resp *gh.Response
			var err error
			if group.Team == "" {
				memberListOpts := &gh.ListMembersOptions{
					ListOptions: listOpts,
				}
				if group.adminsOnly {
					memberListOpts.Role = "admin"
				}
				members, resp, err = cli.Organizations.ListMembers(ctx, owner, memberListOpts)
			} else {
				memberListOpts := &gh.TeamListTeamMembersOptions{
					ListOptions: listOpts,
				}
				members, resp, err = cli.Teams.ListTeamMembersBySlug(ctx, owner, group.Team, memberListOpts)
			}
			if err != nil {
				return userIDs, errors.Wrap(err, "list users for group")
			}
			for _, u := range members {
				// Add results to both group (for persistence) and permissions for user
				accountID := extsvc.AccountID(strconv.FormatInt(u.GetID(), 10))
				group.Users = append(group.Users, accountID)
				addUserToRepoPerms(accountID)
			}
			if resp.NextPage == 0 {
				break
			}
			listOpts.Page = resp.NextPage
		}

		// Persist group
		p.groupsCache.setGroup(group.cachedGroup)
	}

	return userIDs, nil
}

// getUserAffiliatedGroups retrieves affiliated organizations and teams for the given client
// with token. Returned groups are populated from cache if a valid value is available.
//
// 🚨 SECURITY: clientWithToken must be authenticated with a user token.
func (p *Provider) getUserAffiliatedGroups(ctx context.Context, cli *gh.Client, opts authz.FetchPermsOptions) ([]cachedGroup, error) {
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

	// Get orgs
	listOpts := &gh.ListOptions{
		PerPage: 100,
		Page:    1,
	}
	for {
		orgs, resp, err := cli.Organizations.List(ctx, "", listOpts)
		if err != nil {
			return groups, err
		}

		for _, org := range orgs {
			membership, _, err := cli.Organizations.GetOrgMembership(ctx, "", org.GetLogin())
			if err != nil {
				return groups, err
			}

			if canViewOrgRepos(org, membership) {
				syncGroup(org.GetLogin(), "")
			}
		}

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	// Get teams
	listOpts = &gh.ListOptions{
		PerPage: 100,
		Page:    1,
	}
	for {
		teams, resp, err := cli.Teams.ListUserTeams(ctx, listOpts)
		if err != nil {
			return groups, err
		}

		for _, team := range teams {
			if team.GetReposCount() > 0 && team.Organization != nil {
				syncGroup(team.Organization.GetLogin(), team.GetSlug())
			}
		}

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	return groups, nil
}

type repoAffiliatedGroup struct {
	cachedGroup
	// Whether this affiliation is an admin-only affiliation rather than a group-wide
	// affiliation - affects how a sync is conducted.
	adminsOnly bool
}

// getRepoAffiliatedGroups retrieves affiliated organizations and teams for the given repository.
// Returned groups are populated from cache if a valid value is available.
func (p *Provider) getRepoAffiliatedGroups(ctx context.Context, owner, name string, cli *gh.Client, opts authz.FetchPermsOptions) (groups []repoAffiliatedGroup, err error) {
	// Check if repo belongs in an org
	org, resp, err := cli.Organizations.Get(ctx, owner)
	if err != nil {
		if github.IsNotFound(err) || resp.StatusCode == 404 {
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
		groups = append(groups, repoAffiliatedGroup{cachedGroup: group, adminsOnly: adminsOnly})
	}

	// If this repo is an internal repo, we want to allow everyone in the org to read this repo
	// (provided the temporary feature flag is set) irrespective of the user being an admin or not.
	isRepoInternallyVisible := false

	// The visibility field on a repo is only returned if this feature flag is set. As a result
	// there's no point in making an extra API call if this feature flag is not set explicitly.
	if p.enableGithubInternalRepoVisibility {
		var r *gh.Repository
		r, _, err = cli.Repositories.Get(ctx, owner, name)
		if err != nil {
			// Maybe the repo doesn't belong to this org? Or Another error occurred in trying to get the
			// repo. Either way, we are not going to syncGroup for this repo.
			return
		}

		if org != nil && r.GetVisibility() == string(github.VisibilityInternal) {
			isRepoInternallyVisible = true
		}
	}

	allOrgMembersCanRead := isRepoInternallyVisible || canViewOrgRepos(org, nil)
	if allOrgMembersCanRead {
		// 🚨 SECURITY: Iff all members of this org can view this repo, indicate that all members should
		// be sync'd.
		syncGroup(owner, "", false)
	} else {
		// 🚨 SECURITY: Sync *only admins* of this org
		syncGroup(owner, "", true)
		// Also check for teams involved in repo, and indicate all groups should be sync'd.
		listOpts := &gh.ListOptions{
			PerPage: 100,
			Page:    1,
		}
		for {
			var teams []*gh.Team
			var resp *gh.Response
			teams, resp, err = cli.Repositories.ListTeams(ctx, owner, name, listOpts)
			if err != nil {
				return
			}
			for _, t := range teams {
				syncGroup(owner, t.GetSlug(), false)
			}
			if resp.NextPage == 0 {
				break
			}
			listOpts.Page = resp.NextPage
		}
	}

	return groups, nil
}
