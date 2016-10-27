package backend

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mattbaird/gochimp"
	"github.com/sourcegraph/go-github/github"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	store "sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
)

var Orgs = &orgs{}

type orgs struct{}

func (s *orgs) ListOrgs(ctx context.Context, org *sourcegraph.OrgListOptions) (res *sourcegraph.OrgsList, err error) {
	if Mocks.Orgs.ListOrgs != nil {
		return Mocks.Orgs.ListOrgs(ctx, org)
	}

	ctx, done := trace(ctx, "Orgs", "ListOrgs", org, &err)
	defer done()

	client, err := authedGitHubClientC(ctx)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return &sourcegraph.OrgsList{}, nil
	}

	orgs, _, err := client.Organizations.List("", nil)
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

func (s *orgs) ListOrgMembers(ctx context.Context, org *sourcegraph.OrgListOptions) (res *sourcegraph.OrgMembersList, err error) {
	if Mocks.Orgs.ListOrgMembers != nil {
		return Mocks.Orgs.ListOrgMembers(ctx, org)
	}

	ctx, done := trace(ctx, "Orgs", "ListOrgMembers", org, &err)
	defer done()

	client, err := authedGitHubClientC(ctx)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return &sourcegraph.OrgMembersList{}, nil
	}

	// Fetch members of the organization.
	members, _, err := client.Organizations.ListMembers(org.OrgName, &github.ListMembersOptions{})
	if err != nil {
		return nil, err
	}

	var ghIDs []string

	for _, member := range members {
		ghIDs = append(ghIDs, strconv.Itoa(*member.ID))
	}

	// Fetch members of org to see who is on Sourcegraph.
	rUsers, err := authpkg.ListUsersByGitHubID(ctx, ghIDs)
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

// sendEmail lets us avoid sending emails in tests.
var sendEmail = func(template, name, email, subject string, templateContent []gochimp.Var, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
	if notif.EmailIsConfigured() {
		return notif.SendMandrillTemplateBlocking(template, name, email, subject, templateContent, mergeVars)
	}
	return nil, errors.New("email client is not configured")
}

func (s *orgs) InviteUser(ctx context.Context, opt *sourcegraph.UserInvite) (*sourcegraph.UserInviteResponse, error) {
	user := authpkg.ActorFromContext(ctx).User()
	if opt.UserEmail != "" && user != nil {
		_, err := sendEmail("invite-user", opt.UserID, opt.UserEmail, user.Login+" invited you to join "+opt.OrgName+" on Sourcegraph", nil,
			[]gochimp.Var{gochimp.Var{Name: "INVITE_USER", Content: "sourcegraph.com/settings"}, {Name: "FROM_AVATAR", Content: user.AvatarURL}, {Name: "ORG", Content: opt.OrgName}, {Name: "FNAME", Content: user.Login}, {Name: "INVITE_LINK", Content: "https://sourcegraph.com?_event=EmailInviteClicked&_invited_by_user=" + user.Login + "&_org_invite=" + opt.OrgName}})
		if err != nil {
			return nil, fmt.Errorf("Error sending email: %s", err)
		}
	}
	if opt.UserEmail == "" {
		return nil, errors.New("Missing aruments, cannot store.")
	}

	ts := time.Now()
	err := store.UserInvites.Create(ctx, &sourcegraph.UserInvite{
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

// authedGitHubClient returns a new GitHub client that is authenticated using the credentials of the
// context's actor, or nil client if there is no actor (or if the actor has no stored GitHub credentials).
// It returns an error if there was an unexpected error.
func authedGitHubClientC(ctx context.Context) (*github.Client, error) {
	a := authpkg.ActorFromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, nil
	}
	if a.GitHubToken == "" {
		return nil, nil
	}
	ghConf := *githubutil.Default
	ghConf.Context = ctx
	return ghConf.AuthedClient(a.GitHubToken), nil
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
