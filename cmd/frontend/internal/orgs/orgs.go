package orgs

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mattbaird/gochimp"
	"github.com/sourcegraph/go-github/github"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

func ListOrgsPage(ctx context.Context, org *sourcegraph.OrgListOptions) (res *sourcegraph.OrgsList, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return &sourcegraph.OrgsList{}, nil
	}
	client := extgithub.Client(ctx)

	opts := &github.ListOptions{
		Page:    int(org.Page),
		PerPage: int(org.PerPage),
	}
	orgs, _, err := client.Organizations.List("", opts)
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

// ListAllOrgs is a convenience wrapper around ListOrgs (since GitHub's API is paginated)
//
// This method may return an error and a partial list of organizations
func ListAllOrgs(ctx context.Context, op *sourcegraph.OrgListOptions) (res *sourcegraph.OrgsList, err error) {
	// Get a maximum of 1000 organizations per user
	const perPage = 100
	const maxPage = 10
	opts := *op
	opts.PerPage = perPage

	var allOrgs []*sourcegraph.Org
	for page := 1; page <= maxPage; page++ {
		opts.Page = int32(page)
		orgs, err := ListOrgsPage(ctx, &opts)
		if err != nil {
			// If an error occurs, return that error, as well as a list of all organizations
			// collected so far
			return &sourcegraph.OrgsList{
				Orgs: allOrgs}, err
		}
		allOrgs = append(allOrgs, orgs.Orgs...)
		if len(orgs.Orgs) < perPage {
			break
		}
	}
	return &sourcegraph.OrgsList{
		Orgs: allOrgs}, nil
}

func listOrgMembersPage(ctx context.Context, org *sourcegraph.OrgListOptions) (res []*github.User, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return []*github.User{}, nil
	}
	client := extgithub.Client(ctx)

	opts := &github.ListMembersOptions{
		ListOptions: github.ListOptions{
			Page:    int(org.Page),
			PerPage: int(org.PerPage),
		},
	}
	// Fetch members of the organization.
	members, _, err := client.Organizations.ListMembers(org.OrgName, opts)
	if err != nil {
		return nil, err
	}

	return members, nil
}

func ListOrgMembersForInvites(ctx context.Context, org *sourcegraph.OrgListOptions) (res *sourcegraph.OrgMembersList, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return &sourcegraph.OrgMembersList{}, nil
	}
	client := extgithub.Client(ctx)

	members, err := listOrgMembersPage(ctx, org)
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
			orgInvite, _ := store.UserInvites.GetByURI(ctx, *member.Login+org.OrgID)
			if orgInvite == nil || time.Now().Unix()-orgInvite.SentAt.Unix() > 259200 {
				orgMember.CanInvite = true
				fullGHUser, _, err := client.Users.Get(*member.Login)
				if err != nil {
					break
				}
				if fullGHUser.Email != nil {
					orgMember.Email = *fullGHUser.Email
				}
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

// ListAllOrgMembers is a convenience wrapper around ListOrgMembers (since GitHub's API is paginated)
//
// This method may return an error and a partial list of organization members
func ListAllOrgMembers(ctx context.Context, op *sourcegraph.OrgListOptions) (res []*github.User, err error) {
	// Get a maximum of 1,000 members per org
	const perPage = 100
	const maxPage = 10
	opts := *op
	opts.PerPage = perPage

	var allMembers []*github.User
	for page := 1; page <= maxPage; page++ {
		opts.Page = int32(page)
		members, err := listOrgMembersPage(ctx, &opts)
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

func IsOrgMember(ctx context.Context, org *sourcegraph.OrgListOptions) (res bool, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return false, nil
	}
	client := extgithub.Client(ctx)

	isMember, _, err := client.Organizations.IsMember(org.OrgName, org.Username)
	if err != nil {
		return false, err
	}

	return isMember, nil
}

// sendEmail lets us avoid sending emails in tests.
var sendEmail = func(template, name, email, subject string, templateContent []gochimp.Var, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
	if notif.EmailIsConfigured() {
		return notif.SendMandrillTemplateBlocking(template, name, email, subject, templateContent, mergeVars)
	}
	return nil, errors.New("email client is not configured")
}

func InviteUser(ctx context.Context, opt *sourcegraph.UserInvite) (*sourcegraph.UserInviteResponse, error) {
	user := actor.FromContext(ctx).User()
	inviterOrgOptions := &sourcegraph.OrgListOptions{
		OrgName:  opt.OrgName,
		Username: user.Login,
	}
	isInviterMember, err := IsOrgMember(ctx, inviterOrgOptions)
	if err != nil {
		return nil, err
	}
	if !isInviterMember {
		return nil, fmt.Errorf("error sending email: inviting user is not part of organization %s", opt.OrgName)
	}
	inviteeOrgOptions := &sourcegraph.OrgListOptions{
		OrgName:  opt.OrgName,
		Username: opt.UserID,
	}
	isInviteeMember, err := IsOrgMember(ctx, inviteeOrgOptions)
	if err != nil {
		return nil, err
	}
	if !isInviteeMember {
		return nil, fmt.Errorf("error sending email: invited user is not a member of %s", opt.OrgName)
	}

	if opt.UserEmail != "" && user != nil {
		_, err := sendEmail("invite-user", opt.UserID, opt.UserEmail, user.Login+" invited you to join "+opt.OrgName+" on Sourcegraph", nil,
			[]gochimp.Var{gochimp.Var{Name: "INVITE_USER", Content: "sourcegraph.com/settings"}, {Name: "FROM_AVATAR", Content: user.AvatarURL}, {Name: "ORG", Content: opt.OrgName}, {Name: "FNAME", Content: user.Login}, {Name: "INVITE_LINK", Content: "https://sourcegraph.com?_event=EmailInviteClicked&_invited_by_user=" + user.Login + "&_org_invite=" + opt.OrgName}})
		if err != nil {
			return nil, fmt.Errorf("Error sending email: %s", err)
		}
	}
	if opt.UserEmail == "" {
		return nil, errors.New("missing aruments, cannot store")
	}

	ts := time.Now()
	err = store.UserInvites.Create(ctx, &sourcegraph.UserInvite{
		URI:       opt.UserID + opt.OrgID,
		UserID:    opt.UserID,
		UserEmail: opt.UserEmail,
		OrgID:     opt.OrgID,
		OrgName:   opt.OrgName,
		SentAt:    &ts,
	})
	if err != nil {
		return nil, err
	}

	return &sourcegraph.UserInviteResponse{
		OrgName: opt.OrgName,
		OrgID:   opt.OrgID,
	}, nil
}

func toOrg(ghOrg *github.Organization) *sourcegraph.Org {
	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	org := sourcegraph.Org{
		Login:       *ghOrg.Login,
		ID:          int32(*ghOrg.ID),
		AvatarURL:   strv(ghOrg.AvatarURL),
		Name:        strv(ghOrg.Name),
		Blog:        strv(ghOrg.Blog),
		Location:    strv(ghOrg.Location),
		Email:       strv(ghOrg.Email),
		Description: strv(ghOrg.Description)}

	return &org
}

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
