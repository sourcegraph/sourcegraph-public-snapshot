package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/sourcegraph/go-github/github"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

type organizationMemberResolver struct {
	member *sourcegraph.OrgMember
}

func (m *organizationMemberResolver) Login() string {
	return m.member.Login
}

func (m *organizationMemberResolver) GithubID() int32 {
	return int32(m.member.ID)
}

func (m *organizationMemberResolver) Email() string {
	return m.member.Email
}

func (m *organizationMemberResolver) AvatarURL() string {
	return m.member.AvatarURL
}

func (m *organizationMemberResolver) IsSourcegraphUser() bool {
	return m.member.SourcegraphUser
}

// listOrgMembersPage returns a single page of an organization's members
func listOrgMembersPage(ctx context.Context, orgLogin string, opt *sourcegraph.ListOptions) (res []*github.User, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return []*github.User{}, nil
	}

	optGh := &github.ListMembersOptions{
		ListOptions: github.ListOptions{
			Page:    int(opt.Page),
			PerPage: int(opt.PerPage),
		},
	}

	cl := extgithub.Client(ctx)
	installs, _, err := cl.Users.ListInstallations(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, ins := range installs {
		if *ins.Account.Login == orgLogin {
			insClient, err := extgithub.InstallationClient(ctx, *ins.ID)
			if err != nil {
				return nil, err
			}
			members, _, err := insClient.Organizations.ListMembers(ctx, orgLogin, optGh)
			if err != nil {
				return nil, err
			}
			return members, nil
		}
	}
	return nil, fmt.Errorf("github org %s not found for current user", orgLogin)
}

// ListAllOrgMembers is a convenience wrapper around listOrgMembersPage (since GitHub's API is paginated), returning
// a list of all of the specified org's GitHub members
//
// This method may return an error and a partial list of organization members
func ListAllOrgMembers(ctx context.Context, orgLogin string, opt *sourcegraph.ListOptions) (res []*github.User, err error) {
	// Get a maximum of 1,000 members per org
	const perPage = 100
	const maxPage = 10
	opts := *opt
	opts.PerPage = perPage

	var allMembers []*github.User
	for page := 1; page <= maxPage; page++ {
		opts.Page = int32(page)
		members, err := listOrgMembersPage(ctx, orgLogin, &opts)
		if err != nil {
			// If an error occurs, return that error, as well as a list of all members
			// collected so far
			return allMembers, err
		}
		allMembers = append(allMembers, members...)
		if len(members) < perPage {
			break
		}
	}
	return allMembers, nil
}

// IsOrgMember indicates if a user is a member of a given GitHub organization
func IsOrgMember(ctx context.Context, orgLogin string, userLogin string) (res bool, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return false, nil
	}
	client := extgithub.Client(ctx)

	isMember, _, err := client.Organizations.IsMember(ctx, orgLogin, userLogin)
	if err != nil {
		return false, err
	}

	return isMember, nil
}

// toOrgMember converts a GitHub API User object to a Sourcegraph API OrgMember object
func toOrgMember(ghUser *github.User) *sourcegraph.OrgMember {
	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	member := sourcegraph.OrgMember{
		Login:           *ghUser.Login,
		Email:           strv(ghUser.Email),
		ID:              int32(*ghUser.ID),
		AvatarURL:       strv(ghUser.AvatarURL),
		SourcegraphUser: false,
		CanInvite:       true}

	return &member
}
