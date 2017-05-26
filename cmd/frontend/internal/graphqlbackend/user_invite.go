package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mattbaird/gochimp"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	extgithub "sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"
)

type inviteResolver struct {
	invite *sourcegraph.UserInvite
}

func (i *inviteResolver) UserLogin() string {
	return i.invite.UserLogin
}

func (i *inviteResolver) UserEmail() string {
	return i.invite.UserEmail
}

func (i *inviteResolver) OrgLogin() string {
	return i.invite.OrgLogin
}

func (i *inviteResolver) OrgGithubID() (int32, error) {
	v, err := strconv.Atoi(i.invite.OrgID)
	if err != nil {
		return int32(v), nil
	}
	return 0, err
}

func (i *inviteResolver) SentAt() int32 {
	return int32(i.invite.SentAt.Unix())
}

func (i *inviteResolver) URI() string {
	return i.invite.URI
}

func (*schemaResolver) InviteOrgMemberToSourcegraph(ctx context.Context, args *struct {
	OrgLogin    string
	OrgGithubID int32
	UserLogin   string
	UserEmail   string
}) (bool, error) {
	res, err := InviteUser(ctx, &sourcegraph.UserInvite{
		OrgLogin:  args.OrgLogin,
		OrgID:     strconv.Itoa(int(args.OrgGithubID)),
		UserLogin: args.UserLogin,
		UserEmail: args.UserEmail,
	})
	if err != nil {
		return false, err
	}
	if res == sourcegraph.InviteMissingEmail {
		return false, nil
	}
	return true, nil
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
	if isMember, err := IsOrgMember(ctx, opt.OrgLogin, user.Login); err != nil {
		return sourcegraph.InviteError, err
	} else if isMember == false {
		return sourcegraph.InviteError, fmt.Errorf("Error sending email: user %s is not part of organization %s", user.Login, opt.OrgLogin)
	}

	// Confirm invited user is a member of the GitHub organization
	if isMember, err := IsOrgMember(ctx, opt.OrgLogin, opt.UserLogin); err != nil {
		return sourcegraph.InviteError, err
	} else if isMember == false {
		return sourcegraph.InviteError, fmt.Errorf("Error sending email: user %s is not part of organization %s", opt.UserLogin, opt.OrgLogin)
	}

	// If email not provided by frontend, look up this user to see if we can get it
	if opt.UserEmail == "" {
		client := extgithub.Client(ctx)
		invitee, _, err := client.Users.Get(ctx, opt.UserLogin)
		if err != nil {
			return sourcegraph.InviteError, err
		}
		if invitee.Email != nil {
			opt.UserEmail = *invitee.Email
		}
	}

	if opt.UserEmail != "" && user != nil {
		_, err := sendEmail("invite-user", opt.UserLogin, opt.UserEmail, user.Login+" invited you to join "+opt.OrgLogin+" on Sourcegraph", nil,
			[]gochimp.Var{gochimp.Var{Name: "INVITE_USER", Content: "sourcegraph.com/settings"}, {Name: "FROM_AVATAR", Content: user.AvatarURL}, {Name: "ORG", Content: opt.OrgLogin}, {Name: "FNAME", Content: user.Login}, {Name: "INVITE_LINK", Content: "https://sourcegraph.com?_event=EmailInviteClicked&_invited_by_user=" + user.Login + "&_org_invite=" + opt.OrgLogin}})
		if err != nil {
			return sourcegraph.InviteError, fmt.Errorf("Error sending email: %s", err)
		}
	} else {
		return sourcegraph.InviteMissingEmail, nil
	}

	ts := time.Now()
	err := store.UserInvites.Create(ctx, &sourcegraph.UserInvite{
		URI:       opt.UserLogin + opt.OrgID,
		UserLogin: opt.UserLogin,
		UserEmail: opt.UserEmail,
		OrgID:     opt.OrgID,
		OrgLogin:  opt.OrgLogin,
		SentAt:    &ts,
	})
	if err != nil {
		return sourcegraph.InviteError, err
	}

	return sourcegraph.InviteSuccess, nil
}
