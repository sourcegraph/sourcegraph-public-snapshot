package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/invite"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/suspiciousnames"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func (r *schemaResolver) Organization(ctx context.Context, args struct{ Name string }) (*orgResolver, error) {
	org, err := db.Orgs.GetByName(ctx, args.Name)
	if err != nil {
		return nil, err
	}
	return &orgResolver{org: org}, nil
}

// Org is DEPRECATED (but still in use by sourcegraph/src). Use Node to look up an org by its
// graphql.ID instead.
func (r *schemaResolver) Org(ctx context.Context, args *struct {
	ID graphql.ID
}) (*orgResolver, error) {
	return orgByID(ctx, args.ID)
}

func orgByID(ctx context.Context, id graphql.ID) (*orgResolver, error) {
	orgID, err := unmarshalOrgID(id)
	if err != nil {
		return nil, err
	}
	return orgByIDInt32(ctx, orgID)
}

func orgByIDInt32(ctx context.Context, orgID int32) (*orgResolver, error) {
	org, err := db.Orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &orgResolver{org}, nil
}

type orgResolver struct {
	org *types.Org
}

func (o *orgResolver) ID() graphql.ID { return marshalOrgID(o.org.ID) }

func marshalOrgID(id int32) graphql.ID { return relay.MarshalID("Org", id) }

func unmarshalOrgID(id graphql.ID) (orgID int32, err error) {
	err = relay.UnmarshalSpec(id, &orgID)
	return
}

func (o *orgResolver) OrgID() int32 {
	return o.org.ID
}

func (o *orgResolver) Name() string {
	return o.org.Name
}

func (o *orgResolver) DisplayName() *string {
	return o.org.DisplayName
}

func (o *orgResolver) CreatedAt() string { return o.org.CreatedAt.Format(time.RFC3339) }

func (o *orgResolver) Members(ctx context.Context) (*staticUserConnectionResolver, error) {
	// 🚨 SECURITY: Only org members can list the org members.
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		if err == backend.ErrNotAnOrgMember {
			return nil, errors.New("must be a member of this organization to view members")
		}
		return nil, err
	}

	memberships, err := db.OrgMembers.GetByOrgID(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}
	users := make([]*types.User, len(memberships))
	for i, membership := range memberships {
		user, err := db.Users.GetByID(ctx, membership.UserID)
		if err != nil {
			return nil, err
		}
		users[i] = user
	}
	return &staticUserConnectionResolver{users: users}, nil
}

func (o *orgResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	// 🚨 SECURITY: Only organization members and site admins may access the settings, because they
	// may contains secrets or other sensitive data.
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		return nil, err
	}

	settings, err := db.Settings.GetLatest(ctx, api.ConfigurationSubject{Org: &o.org.ID})
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{&configurationSubject{org: o}, settings, nil}, nil
}

func (o *orgResolver) Tags(ctx context.Context) ([]*organizationTagResolver, error) {
	// 🚨 SECURITY: Only organization members and site admins may access the tags.
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		return nil, err
	}

	tags, err := db.OrgTags.GetByOrgID(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}
	organizationTagResolvers := []*organizationTagResolver{}
	for _, tag := range tags {
		organizationTagResolvers = append(organizationTagResolvers, &organizationTagResolver{tag})
	}
	return organizationTagResolvers, nil
}

func (o *orgResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err == backend.ErrNotAuthenticated || err == backend.ErrNotAnOrgMember {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (o *orgResolver) ViewerIsMember(ctx context.Context) (bool, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return false, nil
	}
	if _, err := db.OrgMembers.GetByOrgIDAndUserID(ctx, o.org.ID, actor.UID); err != nil {
		if errcode.IsNotFound(err) {
			err = nil
		}
		return false, err
	}
	return true, nil
}

func getOrgSlackWebhookURL(ctx context.Context, id int32) (string, error) {
	settings, err := backend.Configuration.GetForSubject(ctx, api.ConfigurationSubject{Org: &id})
	if err != nil {
		return "", err
	}
	if settings.NotificationsSlack != nil {
		return settings.NotificationsSlack.WebhookURL, nil
	}
	return "", nil
}

// DEPRECATED: use createOrganization instead
func (*schemaResolver) CreateOrg(ctx context.Context, args *struct {
	Name        string
	DisplayName string
}) (*orgResolver, error) {
	var displayName *string
	if args.DisplayName != "" {
		displayName = &args.DisplayName
	}
	return (&schemaResolver{}).CreateOrganization(ctx, &struct {
		Name        string
		DisplayName *string
	}{
		Name:        args.Name,
		DisplayName: displayName,
	})
}

func (*schemaResolver) CreateOrganization(ctx context.Context, args *struct {
	Name        string
	DisplayName *string
}) (*orgResolver, error) {
	currentUser, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}

	if err := suspiciousnames.CheckNameAllowedForUserOrOrganization(args.Name); err != nil {
		return nil, err
	}
	newOrg, err := db.Orgs.Create(ctx, args.Name, args.DisplayName)
	if err != nil {
		return nil, err
	}

	// Add the current user as the first member of the new org.
	_, err = db.OrgMembers.Create(ctx, newOrg.ID, currentUser.SourcegraphID())
	if err != nil {
		return nil, err
	}

	return &orgResolver{org: newOrg}, nil
}

func (*schemaResolver) UpdateOrganization(ctx context.Context, args *struct {
	ID          graphql.ID
	DisplayName *string
}) (*orgResolver, error) {
	var orgID int32
	if err := relay.UnmarshalSpec(args.ID, &orgID); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is a member
	// of the org that is being modified.
	if err := backend.CheckOrgAccess(ctx, orgID); err != nil {
		return nil, err
	}

	updatedOrg, err := db.Orgs.Update(ctx, orgID, args.DisplayName)
	if err != nil {
		return nil, err
	}

	return &orgResolver{org: updatedOrg}, nil
}

// DEPRECATED: use RemoveUserFromOrganization instead.
func (*schemaResolver) RemoveUserFromOrg(ctx context.Context, args *struct {
	UserID graphql.ID
	OrgID  graphql.ID
}) (*EmptyResponse, error) {
	return (&schemaResolver{}).RemoveUserFromOrganization(ctx, &struct {
		User         graphql.ID
		Organization graphql.ID
	}{
		User:         args.UserID,
		Organization: args.OrgID,
	})
}

func (*schemaResolver) RemoveUserFromOrganization(ctx context.Context, args *struct {
	User         graphql.ID
	Organization graphql.ID
}) (*EmptyResponse, error) {
	orgID, err := unmarshalOrgID(args.Organization)
	if err != nil {
		return nil, err
	}
	userID, err := unmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is a member of the org that is being modified, or a
	// site admin.
	if err := backend.CheckOrgAccess(ctx, orgID); err != nil {
		return nil, err
	}

	log15.Info("removing user from org", "user", userID, "org", orgID)
	return nil, db.OrgMembers.Remove(ctx, orgID, userID)
}

func getUserToInviteToOrganization(ctx context.Context, username string, orgID int32) (userToInvite *types.User, userEmailAddress string, err error) {
	userToInvite, err = db.Users.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", err
	}

	if conf.CanSendEmail() {
		// Look up user's email address so we can send them an email (if needed).
		email, verified, err := db.UserEmails.GetPrimaryEmail(ctx, userToInvite.ID)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, "", errors.WithMessage(err, "looking up invited user's primary email address")
		}
		if verified {
			// Completely discard unverified emails.
			userEmailAddress = email
		}
	}

	if _, err := db.OrgMembers.GetByOrgIDAndUserID(ctx, orgID, userToInvite.ID); err == nil {
		return nil, "", errors.New("user is already a member of the organization")
	} else if _, ok := err.(*db.ErrOrgMemberNotFound); !ok {
		return nil, "", err
	}
	return userToInvite, userEmailAddress, nil
}

type inviteUserToOrganizationResult struct {
	sentInvitationEmail bool
	acceptInvitationURL string
}

func (r *inviteUserToOrganizationResult) SentInvitationEmail() bool   { return r.sentInvitationEmail }
func (r *inviteUserToOrganizationResult) AcceptInvitationURL() string { return r.acceptInvitationURL }

func (*schemaResolver) InviteUserToOrganization(ctx context.Context, args *struct {
	Organization graphql.ID
	Username     string
}) (*inviteUserToOrganizationResult, error) {
	var orgID int32
	if err := relay.UnmarshalSpec(args.Organization, &orgID); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Check that the current user is a member of the org that the user is being
	// invited to.
	if err := backend.CheckOrgAccess(ctx, orgID); err != nil {
		return nil, err
	}

	currentUser, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("must be logged in")
	}
	senderEmail, senderEmailVerified, err := db.UserEmails.GetPrimaryEmail(ctx, currentUser.SourcegraphID())
	if err != nil {
		return nil, err
	}

	_, recipientEmail, err := getUserToInviteToOrganization(ctx, args.Username, orgID)
	if err != nil {
		return nil, err
	}

	if envvar.SourcegraphDotComMode() {
		// Only allow email-verified users to send invites.
		if !senderEmailVerified {
			return nil, errors.New("must verify your email to send invites")
		}

		// Check and decrement our invite quota, to prevent abuse (sending too many invites).
		//
		// There is no user invite quota for on-prem instances because we assume they can
		// trust their users to not abuse invites.
		if ok, err := db.Users.CheckAndDecrementInviteQuota(ctx, currentUser.SourcegraphID()); err != nil {
			return nil, err
		} else if !ok {
			return nil, errors.New("invite quota exceeded (contact support to increase the quota)")
		}
	}

	org, err := db.Orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	token, err := invite.CreateOrgToken(recipientEmail, org)
	if err != nil {
		return nil, err
	}

	inviteURL := globals.AppURL.String() + "/settings/accept-invite?token=" + token

	if conf.CanSendEmail() && recipientEmail != "" {
		// If email is disabled, the frontend will still show the invitation link.
		var fromName string
		if currentUser.user.DisplayName != "" {
			fromName = fmt.Sprintf("%s (%s on Sourcegraph)", currentUser.user.DisplayName, currentUser.user.Username)
		} else {
			fromName = fmt.Sprintf("%s (on Sourcegraph)", currentUser.user.Username)
		}
		if err := invite.SendEmail(recipientEmail, fromName, org.Name, inviteURL); err != nil {
			return nil, err
		}
	}

	slackWebhookURL, err := getOrgSlackWebhookURL(ctx, org.ID)
	if err != nil {
		return nil, err
	}
	client := slack.New(slackWebhookURL, true)
	go slack.NotifyOnInvite(client, currentUser, senderEmail, org, recipientEmail)

	return &inviteUserToOrganizationResult{
		sentInvitationEmail: recipientEmail != "",
		acceptInvitationURL: inviteURL,
	}, nil
}

func (*schemaResolver) AcceptUserInvite(ctx context.Context, args *struct {
	InviteToken string
}) (*EmptyResponse, error) {
	currentUser, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}
	email, _, err := db.UserEmails.GetPrimaryEmail(ctx, currentUser.SourcegraphID())
	if err != nil {
		return nil, err
	}

	token, err := invite.ParseToken(args.InviteToken)
	if err != nil {
		return nil, err
	}
	org, err := db.Orgs.GetByID(ctx, token.OrgID)
	if err != nil {
		return nil, err
	}

	_, err = db.OrgMembers.Create(ctx, token.OrgID, currentUser.SourcegraphID())
	if err != nil {
		return nil, err
	}

	slackWebhookURL, err := getOrgSlackWebhookURL(ctx, org.ID)
	if err != nil {
		return nil, err
	}
	client := slack.New(slackWebhookURL, true)
	go slack.NotifyOnAcceptedInvite(client, currentUser, email, org)

	return &EmptyResponse{}, nil
}

func (*schemaResolver) AddUserToOrganization(ctx context.Context, args *struct {
	Organization graphql.ID
	Username     string
}) (*EmptyResponse, error) {
	var orgID int32
	if err := relay.UnmarshalSpec(args.Organization, &orgID); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Must be a site admin to immediately add a user to an organization (bypassing the
	// invitation step).
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	userToInvite, _, err := getUserToInviteToOrganization(ctx, args.Username, orgID)
	if err != nil {
		return nil, err
	}
	if _, err := db.OrgMembers.Create(ctx, orgID, userToInvite.ID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
