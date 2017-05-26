package graphqlbackend

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/go-github/github"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
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

func (m *organizationMemberResolver) CanInvite() bool {
	return m.member.CanInvite
}

func (m *organizationMemberResolver) Invite() *inviteResolver {
	if m.member.Invite == nil {
		return nil
	}
	return &inviteResolver{invite: m.member.Invite}
}

// listOrgMembersPage returns a single page of an organization's members
func listOrgMembersPage(ctx context.Context, orgLogin string, opt *sourcegraph.ListOptions) (res []*github.User, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return []*github.User{}, nil
	}
	client := extgithub.Client(ctx)

	optGh := &github.ListMembersOptions{
		ListOptions: github.ListOptions{
			Page:    int(opt.Page),
			PerPage: int(opt.PerPage),
		},
	}
	// Fetch members of the organization.
	members, _, err := client.Organizations.ListMembers(orgLogin, optGh)
	if err != nil {
		return nil, err
	}

	return members, nil
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

// ListOrgMembersForInvites returns a list of org members with context required to invite them to Sourcegraph
// TODO: make this function capable of returning more than a single page of org members
func ListOrgMembersForInvites(ctx context.Context, orgLogin string, orgID int, opt *sourcegraph.ListOptions) (res *sourcegraph.OrgMembersList, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return &sourcegraph.OrgMembersList{}, nil
	}

	members, err := listOrgMembersPage(ctx, orgLogin, opt)
	if err != nil {
		return nil, err
	}

	var ghIDs []string

	for _, member := range members {
		ghIDs = append(ghIDs, strconv.Itoa(*member.ID))
	}

	// Fetch members of org to see who is on Sourcegraph.
	rUsers, err := auth0.ListUsersByGitHubID(ctx, ghIDs)
	if err != nil {
		return nil, err
	}

	slice := []*sourcegraph.OrgMember{}
	// Disable inviting for members on Sourcegraph and fetch full GitHub info for public email on non-registered Sourcegraph users
	for _, member := range members {
		orgMember := toOrgMember(member)
		if rUser, ok := rUsers[strconv.Itoa(*member.ID)]; ok {
			orgMember.SourcegraphUser = true
			orgMember.CanInvite = false
			orgMember.Email = rUser.Email
		} else {
			orgInvite, _ := store.UserInvites.GetByURI(ctx, *member.Login+strconv.Itoa(orgID))

			if orgInvite == nil || time.Now().Unix()-orgInvite.SentAt.Unix() > 259200 {
				orgMember.CanInvite = true
				// Emails are not available through the GH API for org members. So instead of eagerly fetching each member's email,
				// we defer, and do lazy fetching only when the user intends to invite someone.
				orgMember.Email = ""
			} else {
				orgMember.CanInvite = false
				orgMember.Invite = orgInvite
				orgMember.Email = orgInvite.UserEmail
			}
		}

		slice = append(slice, orgMember)
	}

	return &sourcegraph.OrgMembersList{
		OrgMembers: slice}, nil
}

// IsOrgMember indicates if a user is a member of a given GitHub organization
func IsOrgMember(ctx context.Context, orgLogin string, userLogin string) (res bool, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return false, nil
	}
	client := extgithub.Client(ctx)

	isMember, _, err := client.Organizations.IsMember(orgLogin, userLogin)
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
