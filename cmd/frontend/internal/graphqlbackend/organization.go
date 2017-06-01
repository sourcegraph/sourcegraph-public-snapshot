package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/go-github/github"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

type organizationResolver struct {
	organization *sourcegraph.Org
}

func (o *organizationResolver) Login() string {
	return o.organization.Login
}

func (o *organizationResolver) GithubID() int32 {
	return int32(o.organization.ID)
}

func (o *organizationResolver) Email() string {
	return o.organization.Email
}

func (o *organizationResolver) Name() string {
	return o.organization.Name
}

func (o *organizationResolver) AvatarURL() string {
	return o.organization.AvatarURL
}

func (o *organizationResolver) Description() string {
	return o.organization.Description
}

func (o *organizationResolver) Collaborators() int32 {
	return o.organization.Collaborators
}

func (o *organizationResolver) Members(ctx context.Context) ([]*organizationMemberResolver, error) {
	// TODO(Dan): this method currently only returns a single page of results
	membersList, err := ListOrgMembersForInvites(ctx, o.organization.Login, int(o.organization.ID), &sourcegraph.ListOptions{})
	if err != nil {
		return nil, err
	}

	res := make([]*organizationMemberResolver, len(membersList.OrgMembers))
	for i, member := range membersList.OrgMembers {
		res[i] = &organizationMemberResolver{member}
	}
	return res, nil
}

// GetOrg returns a single *sourcegraph.Org representing a single GitHub organization
func GetOrg(ctx context.Context, orgName string) (*sourcegraph.Org, error) {
	client := extgithub.Client(ctx)

	org, _, err := client.Organizations.Get(ctx, orgName)
	if err != nil {
		return nil, err
	}

	return toOrg(org), nil
}

// listOrgsPage returns a single page of the current user's GitHub organizations
func listOrgsPage(ctx context.Context, opt *sourcegraph.ListOptions) (res *sourcegraph.OrgsList, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return &sourcegraph.OrgsList{}, nil
	}
	client := extgithub.Client(ctx)

	orgs, _, err := client.Organizations.List(ctx, "", &github.ListOptions{
		Page:    int(opt.Page),
		PerPage: int(opt.PerPage),
	})
	if err != nil {
		return nil, err
	}

	slice := []*sourcegraph.Org{}
	for _, k := range orgs {
		slice = append(slice, toOrg(k))
	}

	return &sourcegraph.OrgsList{
		Orgs: slice}, nil
}

// ListAllOrgs is a convenience wrapper around listOrgsPage (since GitHub's API is paginated), returning
// a list of all of the current user's GitHub organizations
//
// This method may return an error and a partial list of organizations
func ListAllOrgs(ctx context.Context, op *sourcegraph.ListOptions) (res *sourcegraph.OrgsList, err error) {
	if feature.Features.GitHubApps {
		var orgs []*sourcegraph.Org
		if !extgithub.HasAuthedUser(ctx) {
			return &sourcegraph.OrgsList{}, nil
		}
		cl := extgithub.Client(ctx)
		installs, _, err := cl.Users.ListInstallations(ctx, nil)
		if err != nil {
			return nil, err
		}
		for _, ins := range installs {
			orgs = append(orgs, toOrgFromAccount(ins.Account))
		}
		return &sourcegraph.OrgsList{Orgs: orgs}, nil
	}

	// Get a maximum of 1000 organizations per user
	const perPage = 100
	const maxPage = 10
	opts := *op
	opts.PerPage = perPage

	var allOrgs []*sourcegraph.Org
	for page := 1; page <= maxPage; page++ {
		opts.Page = int32(page)
		orgsPage, err := listOrgsPage(ctx, &opts)
		if err != nil {
			// If an error occurs, return that error, as well as a list of all organizations
			// collected so far
			return &sourcegraph.OrgsList{
				Orgs: allOrgs}, err
		}
		allOrgs = append(allOrgs, orgsPage.Orgs...)
		if len(orgsPage.Orgs) < perPage {
			break
		}
	}
	return &sourcegraph.OrgsList{Orgs: allOrgs}, nil
}

func OrganizationRepos(ctx context.Context, org string, opt *github.RepositoryListByOrgOptions) ([]*github.Repository, error) {
	repo, _, err := extgithub.Client(ctx).Repositories.ListByOrg(ctx, org, opt)
	return repo, err
}

// toOrg converts a GitHub API Organization object to a Sourcegraph API Org object
func toOrg(ghOrg *github.Organization) *sourcegraph.Org {
	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	intv := func(i *int) int32 {
		if i == nil {
			return 0
		}
		return int32(*i)
	}

	org := sourcegraph.Org{
		Login:         *ghOrg.Login,
		ID:            int32(*ghOrg.ID),
		AvatarURL:     strv(ghOrg.AvatarURL),
		Name:          strv(ghOrg.Name),
		Blog:          strv(ghOrg.Blog),
		Location:      strv(ghOrg.Location),
		Email:         strv(ghOrg.Email),
		Description:   strv(ghOrg.Description),
		Collaborators: intv(ghOrg.Collaborators)}

	return &org
}

func toOrgFromAccount(user *github.User) *sourcegraph.Org {
	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	intv := func(i *int) int32 {
		if i == nil {
			return 0
		}
		return int32(*i)
	}

	org := sourcegraph.Org{
		Login:         *user.Login,
		ID:            int32(*user.ID),
		AvatarURL:     strv(user.AvatarURL),
		Name:          strv(user.Name),
		Blog:          strv(user.Blog),
		Location:      strv(user.Location),
		Email:         strv(user.Email),
		Description:   "",
		Collaborators: intv(user.Collaborators),
	}

	return &org
}
