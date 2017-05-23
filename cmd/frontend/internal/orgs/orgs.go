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

// GetOrg returns a single *sourcegraph.Org representing a single GitHub organization
func GetOrg(ctx context.Context, orgName string) (*sourcegraph.Org, error) {
	client := extgithub.Client(ctx)

	org, _, err := client.Organizations.Get(orgName)
	if err != nil {
		return nil, err
	}

	return toOrg(org), nil
}

// ListOrgsPage returns a *sourcegraph.OrgList represnting a single "page" of GitHub's organizations API
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

// ListOrgMembersForInvites returns a list of org members with context required to invite them to Sourcegraph
// TODO: make this function capable of returning more than a single page of org members
func ListOrgMembersForInvites(ctx context.Context, org *sourcegraph.OrgListOptions) (res *sourcegraph.OrgMembersList, err error) {
	if !extgithub.HasAuthedUser(ctx) {
		return &sourcegraph.OrgMembersList{}, nil
	}

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

// InviteUser invites a member of an organization to Sourcegraph
// This function adds the invitation details to a Postgres database and sends the target an invitation email
// through Mandrill
func InviteUser(ctx context.Context, opt *sourcegraph.UserInvite) (sourcegraph.UserInviteResponse, error) {
	user := actor.FromContext(ctx).User()
	if user == nil {
		return sourcegraph.InviteError, errors.New("Inviting user is not signed in")
	}

	// Confirm inviting usre is a member of the GitHub organization
	err := validateMembership(ctx, opt.OrgName, user.Login)
	if err != nil {
		return sourcegraph.InviteError, err
	}

	// Confirm invited user is a member of the GitHub organization
	err = validateMembership(ctx, opt.OrgName, opt.UserID)
	if err != nil {
		return sourcegraph.InviteError, err
	}

	// If email not provided by frontend, look up this user to see if we can get it
	if opt.UserEmail == "" {
		client := extgithub.Client(ctx)
		invitee, _, err := client.Users.Get(opt.UserID)
		if err != nil {
			return sourcegraph.InviteError, err
		}
		if invitee.Email != nil {
			opt.UserEmail = *invitee.Email
		}
	}

	if opt.UserEmail != "" && user != nil {
		_, err := sendEmail("invite-user", opt.UserID, opt.UserEmail, user.Login+" invited you to join "+opt.OrgName+" on Sourcegraph", nil,
			[]gochimp.Var{gochimp.Var{Name: "INVITE_USER", Content: "sourcegraph.com/settings"}, {Name: "FROM_AVATAR", Content: user.AvatarURL}, {Name: "ORG", Content: opt.OrgName}, {Name: "FNAME", Content: user.Login}, {Name: "INVITE_LINK", Content: "https://sourcegraph.com?_event=EmailInviteClicked&_invited_by_user=" + user.Login + "&_org_invite=" + opt.OrgName}})
		if err != nil {
			return sourcegraph.InviteError, fmt.Errorf("Error sending email: %s", err)
		}
	} else {
		return sourcegraph.InviteMissingEmail, nil
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
		return sourcegraph.InviteError, err
	}

	return sourcegraph.InviteSuccess, nil
}

// validateMembership validates that a given GitHub user is a member of a given GitHub organization
func validateMembership(ctx context.Context, orgName string, userID string) error {
	inviterOrgOptions := &sourcegraph.OrgListOptions{
		OrgName:  orgName,
		Username: userID,
	}
	isInviterMember, err := IsOrgMember(ctx, inviterOrgOptions)
	if err != nil {
		return err
	}
	if !isInviterMember {
		return fmt.Errorf("Error sending email: user %s is not part of organization %s", userID, orgName)
	}
	return nil
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
