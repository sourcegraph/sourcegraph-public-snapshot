package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/invite"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
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

func (o *orgResolver) Threads(ctx context.Context, args *struct {
	RepoCanonicalRemoteID *string // TODO(nick): deprecated
	CanonicalRemoteIDs    *[]string
	Branch                *string
	File                  *string
	Limit                 *int32
}) (*threadConnectionResolver, error) {
	// 🚨 SECURITY: Only organization members and site admins may access the threads, because they
	// may contain secrets or other sensitive data.
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		return nil, err
	}

	var canonicalRemoteIDs []api.RepoURI
	if args.CanonicalRemoteIDs != nil {
		for _, canonicalRemoteID := range *args.CanonicalRemoteIDs {
			canonicalRemoteIDs = append(canonicalRemoteIDs, api.RepoURI(canonicalRemoteID))
		}
	}
	if args.RepoCanonicalRemoteID != nil {
		canonicalRemoteIDs = append(canonicalRemoteIDs, api.RepoURI(*args.RepoCanonicalRemoteID))
	}
	var repos []*types.OrgRepo
	if len(canonicalRemoteIDs) > 0 {
		var err error
		repos, err = db.OrgRepos.GetByCanonicalRemoteIDs(ctx, o.org.ID, canonicalRemoteIDs)
		if err != nil {
			return nil, err
		}
	}
	return &threadConnectionResolver{o.org, repos, canonicalRemoteIDs, args.File, args.Branch, args.Limit}, nil
}

func (o *orgResolver) Tags(ctx context.Context) ([]*orgTagResolver, error) {
	// 🚨 SECURITY: Only organization members and site admins may access the tags.
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		return nil, err
	}

	tags, err := db.OrgTags.GetByOrgID(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}
	orgTagResolvers := []*orgTagResolver{}
	for _, tag := range tags {
		orgTagResolvers = append(orgTagResolvers, &orgTagResolver{tag})
	}
	return orgTagResolvers, nil
}

func (o *orgResolver) Repo(ctx context.Context, args *struct {
	CanonicalRemoteID string
}) (*orgRepoResolver, error) {
	// 🚨 SECURITY: Only organization members and site admins may access the organization's repositories..
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		return nil, err
	}

	orgRepo, err := getOrgRepo(ctx, o.org.ID, api.RepoURI(args.CanonicalRemoteID))
	if err != nil {
		return nil, err
	}
	return &orgRepoResolver{o.org, orgRepo}, nil
}

func getOrgRepo(ctx context.Context, orgID int32, canonicalRemoteID api.RepoURI) (*types.OrgRepo, error) {
	// 🚨 SECURITY: Only organization members and site admins may access the organization's repositories..
	if err := backend.CheckOrgAccess(ctx, orgID); err != nil {
		return nil, err
	}

	orgRepo, err := db.OrgRepos.GetByCanonicalRemoteID(ctx, orgID, canonicalRemoteID)
	if errcode.IsNotFound(err) {
		// We don't want to create org repos just because an org member queried for threads
		// and we don't want the client to think this is an error.
		err = nil
	}
	return orgRepo, err
}

func (o *orgResolver) Repos(ctx context.Context) ([]*orgRepoResolver, error) {
	// 🚨 SECURITY: Only organization members and site admins may access the organization's repositories..
	if err := backend.CheckOrgAccess(ctx, o.org.ID); err != nil {
		return nil, err
	}

	repos, err := db.OrgRepos.GetByOrg(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}
	orgRepoResolvers := []*orgRepoResolver{}
	for _, repo := range repos {
		orgRepoResolvers = append(orgRepoResolvers, &orgRepoResolver{o.org, repo})
	}
	return orgRepoResolvers, nil
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

func (*schemaResolver) CreateOrg(ctx context.Context, args *struct {
	Name        string
	DisplayName string
}) (*orgResolver, error) {
	currentUser, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
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

func (*schemaResolver) UpdateOrg(ctx context.Context, args *struct {
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

	log15.Info("updating org", "org", args.ID, "display name", args.DisplayName)

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

func getUserToInviteToOrganization(ctx context.Context, usernameOrEmail string, orgID int32) (userToInvite *types.User, userEmailAddress string, err error) {
	// See if the user to invite exists.
	if isEmailAddress := strings.Contains(usernameOrEmail, "@"); isEmailAddress {
		// Invite user by email address.
		var err error
		userToInvite, err = db.Users.GetByVerifiedEmail(ctx, usernameOrEmail)
		if errcode.IsNotFound(err) {
			err = nil // not a fatal error, can send invite link to user
		}
		if err != nil {
			return nil, "", err
		}
		userEmailAddress = usernameOrEmail
	} else {
		// Invite user by username. A user with the given username must exist.
		var err error
		userToInvite, err = db.Users.GetByUsername(ctx, usernameOrEmail)
		if err != nil {
			return nil, "", err
		}

		// Look up user's email address so we can send them an email (if needed).
		var verified bool
		userEmailAddress, verified, err = db.UserEmails.GetPrimaryEmail(ctx, userToInvite.ID)
		if err != nil {
			return nil, "", err
		}
		if !verified && conf.CanSendEmail() {
			return nil, "", errors.New("user has no verified email addresses to send invitation to")
		}
	}

	if userToInvite != nil {
		_, err = db.OrgMembers.GetByOrgIDAndUserID(ctx, orgID, userToInvite.ID)
		if err == nil {
			return nil, "", errors.New("user is already a member of the organization")
		}
		if _, ok := err.(*db.ErrOrgMemberNotFound); !ok {
			return nil, "", err
		}
	}

	return userToInvite, userEmailAddress, nil
}

type inviteUserResult struct {
	acceptInviteURL string
}

func (r *inviteUserResult) AcceptInviteURL() string { return r.acceptInviteURL }

func (*schemaResolver) InviteUserToOrganization(ctx context.Context, args *struct {
	Organization    graphql.ID
	UsernameOrEmail string
}) (*inviteUserResult, error) {
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
	email, emailVerified, err := db.UserEmails.GetPrimaryEmail(ctx, currentUser.SourcegraphID())
	if err != nil {
		return nil, err
	}

	_, userEmailAddress, err := getUserToInviteToOrganization(ctx, args.UsernameOrEmail, orgID)
	if err != nil {
		return nil, err
	}

	if envvar.SourcegraphDotComMode() {
		// Only allow email-verified users to send invites.
		if !emailVerified {
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

	token, err := invite.CreateOrgToken(userEmailAddress, org)
	if err != nil {
		return nil, err
	}

	inviteURL := globals.AppURL.String() + "/settings/accept-invite?token=" + token

	if conf.CanSendEmail() {
		// If email is disabled, the frontend will show a link instead.
		var fromName string
		if currentUser.user.DisplayName != "" {
			fromName = fmt.Sprintf("%s (%s on Sourcegraph)", currentUser.user.DisplayName, currentUser.user.Username)
		} else {
			fromName = fmt.Sprintf("%s (on Sourcegraph)", currentUser.user.Username)
		}
		if err := invite.SendEmail(userEmailAddress, fromName, org.Name, inviteURL); err != nil {
			return nil, err
		}
	}

	slackWebhookURL, err := getOrgSlackWebhookURL(ctx, org.ID)
	if err != nil {
		return nil, err
	}
	client := slack.New(slackWebhookURL, true)
	go slack.NotifyOnInvite(client, currentUser, email, org, userEmailAddress)

	return &inviteUserResult{acceptInviteURL: inviteURL}, nil
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
	Organization    graphql.ID
	UsernameOrEmail string
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

	userToInvite, _, err := getUserToInviteToOrganization(ctx, args.UsernameOrEmail, orgID)
	if err != nil {
		return nil, err
	}
	if userToInvite == nil {
		return nil, errors.New("user does not exist (the user must sign up first before they can be added to an organization, but you can invite users without an account, and they'll be prompted to sign up and join the organization)")
	}

	if _, err := db.OrgMembers.Create(ctx, orgID, userToInvite.ID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

// unmarshalOrgGraphQLID unmarshals and returns the int32 org ID of the first
// non-nil element of ids.
func unmarshalOrgGraphQLID(ids ...*graphql.ID) (int32, error) {
	for _, id := range ids {
		if id != nil {
			var orgID int32
			err := relay.UnmarshalSpec(*id, &orgID)
			return orgID, err
		}
	}
	return 0, errors.New("at least 1 of id and orgID must be specified")
}
