package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/invite"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/slack"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth0"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func (r *rootResolver) Org(ctx context.Context, args *struct {
	ID int32
}) (*orgResolver, error) {
	// 🚨 SECURITY: Check that the current user is a member of the org.
	actor := actor.FromContext(ctx)
	if _, err := localstore.OrgMembers.GetByOrgIDAndUserID(ctx, args.ID, actor.UID); err != nil {
		return nil, err
	}
	org, err := localstore.Orgs.GetByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return &orgResolver{org}, nil
}

type orgResolver struct {
	org *sourcegraph.Org
}

func (o *orgResolver) ID() int32 {
	return o.org.ID
}

func (o *orgResolver) Name() string {
	return o.org.Name
}

func (o *orgResolver) DisplayName() *string {
	return o.org.DisplayName
}

func (o *orgResolver) SlackWebhookURL() *string {
	return o.org.SlackWebhookURL
}

func (o *orgResolver) Members(ctx context.Context) ([]*orgMemberResolver, error) {
	sgMembers, err := store.OrgMembers.GetByOrgID(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}

	members := []*orgMemberResolver{}
	for _, sgMember := range sgMembers {
		member := &orgMemberResolver{o.org, sgMember, nil}
		members = append(members, member)
	}
	return members, nil
}

func (o *orgResolver) LatestSettings(ctx context.Context) (*orgSettingsResolver, error) {
	setting, err := store.OrgSettings.GetLatestByOrgID(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, nil
	}
	return &orgSettingsResolver{o.org, setting, nil}, nil
}

func (o *orgResolver) Threads(ctx context.Context, args *struct {
	RepoRemoteURI         *string // DEPRECATED: use RepoCanonicalRemoteID instead.
	RepoCanonicalRemoteID *string
	Branch                *string
	File                  *string
	Limit                 *int32
}) (*threadConnectionResolver, error) {
	var repo *sourcegraph.OrgRepo
	if args.RepoRemoteURI != nil {
		args.RepoCanonicalRemoteID = args.RepoRemoteURI
	}
	if args.RepoCanonicalRemoteID != nil {
		var err error
		repo, err = getOrgRepo(ctx, o.org.ID, *args.RepoCanonicalRemoteID)
		if err != nil {
			return nil, err
		}
	}
	return &threadConnectionResolver{o.org, repo, args.RepoCanonicalRemoteID, args.File, args.Branch, args.Limit}, nil
}

func (o *orgResolver) Tags(ctx context.Context) ([]*orgTagResolver, error) {
	tags, err := store.OrgTags.GetByOrgID(ctx, o.org.ID)
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
	RemoteURI         *string // DEPRECATED: use CanonicalRemoteID instead.
	CanonicalRemoteID *string
}) (*orgRepoResolver, error) {
	if args.RemoteURI == nil && args.CanonicalRemoteID == nil {
		return nil, errors.New("canonicalRemoteID required")
	}
	if args.RemoteURI != nil {
		args.CanonicalRemoteID = args.RemoteURI
	}
	orgRepo, err := getOrgRepo(ctx, o.org.ID, *args.CanonicalRemoteID)
	if err != nil {
		return nil, err
	}
	return &orgRepoResolver{o.org, orgRepo}, nil
}

func getOrgRepo(ctx context.Context, orgID int32, canonicalRemoteID string) (*sourcegraph.OrgRepo, error) {
	orgRepo, err := store.OrgRepos.GetByCanonicalRemoteID(ctx, orgID, canonicalRemoteID)
	if err == store.ErrRepoNotFound {
		// We don't want to create org repos just because an org member queried for threads
		// and we don't want the client to think this is an error.
		err = nil
	}
	return orgRepo, err
}

func (o *orgResolver) Repos(ctx context.Context) ([]*orgRepoResolver, error) {
	repos, err := store.OrgRepos.GetByOrg(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}
	orgRepoResolvers := []*orgRepoResolver{}
	for _, repo := range repos {
		orgRepoResolvers = append(orgRepoResolvers, &orgRepoResolver{o.org, repo})
	}
	return orgRepoResolvers, nil
}

func (*schemaResolver) CreateOrg(ctx context.Context, args *struct {
	Name        string
	DisplayName string
}) (*orgResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	newOrg, err := store.Orgs.Create(ctx, args.Name, args.DisplayName)
	if err != nil {
		return nil, err
	}

	// Add the current user as the first member of the new org.
	_, err = store.OrgMembers.Create(ctx, newOrg.ID, actor.UID)
	if err != nil {
		return nil, err
	}

	// Add the editor-beta tag to all orgs created
	_, err = store.OrgTags.Create(ctx, newOrg.ID, "editor-beta")
	if err != nil {
		return nil, err
	}

	return &orgResolver{org: newOrg}, nil
}

func (*schemaResolver) UpdateOrg(ctx context.Context, args *struct {
	OrgID           int32
	DisplayName     *string
	SlackWebhookURL *string
}) (*orgResolver, error) {
	// 🚨 SECURITY: Check that the current user is a member
	// of the org that is being modified.
	actor := actor.FromContext(ctx)
	if _, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID, actor.UID); err != nil {
		return nil, err
	}
	log15.Info("updating org", "org", args.OrgID, "display name", args.DisplayName, "webhook URL", args.SlackWebhookURL, "actor", actor.UID)

	updatedOrg, err := store.Orgs.Update(ctx, args.OrgID, args.DisplayName, args.SlackWebhookURL)
	if err != nil {
		return nil, err
	}

	return &orgResolver{org: updatedOrg}, nil
}

func (*schemaResolver) RemoveUserFromOrg(ctx context.Context, args *struct {
	UserID string
	OrgID  int32
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Check that the current user is a member
	// of the org that is being modified.
	actor := actor.FromContext(ctx)
	if _, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID, actor.UID); err != nil {
		return nil, err
	}
	log15.Info("removing user from org", "user", args.UserID, "org", args.OrgID)
	return nil, store.OrgMembers.Remove(ctx, args.OrgID, args.UserID)
}

func (*schemaResolver) InviteUser(ctx context.Context, args *struct {
	OrgID int32
	Email string
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Check that the current user is a member
	// of the org that is being modified.
	actor := actor.FromContext(ctx)
	orgMember, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID, actor.UID)
	if err != nil {
		return nil, err
	}

	user, err := store.Users.GetByAuth0ID(ctx, orgMember.UserID)
	if err != nil {
		return nil, err
	}

	// Don't invite the user if they are already a member.
	invitedUser, err := store.Users.GetByEmail(ctx, args.Email)
	if err != nil {
		if _, ok := err.(store.ErrUserNotFound); !ok {
			return nil, err
		}
	}

	if invitedUser != nil {
		_, err = store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID, invitedUser.Auth0ID)
		if err == nil {
			return nil, fmt.Errorf("%s is already a member of org %d", args.Email, args.OrgID)
		}
		if _, ok := err.(store.ErrOrgMemberNotFound); !ok {
			return nil, err
		}
		// Add the editor beta tag to an invited user if they're already registered
		_, err := store.UserTags.CreateIfNotExists(ctx, invitedUser.ID, "editor-beta")
		if err != nil {
			return nil, err
		}
	}

	org, err := localstore.Orgs.GetByID(ctx, args.OrgID)
	if err != nil {
		return nil, err
	}

	token, err := invite.CreateOrgToken(args.Email, org)
	if err != nil {
		return nil, err
	}

	invite.SendEmail(args.Email, user.DisplayName, org.Name, token)
	if err != nil {
		return nil, err
	}

	if user, err := currentUser(ctx); err != nil {
		// errors swallowed because user is only needed for Slack notifications
		log15.Error("graphqlbackend.InviteUser: currentUser failed", "error", err)
	} else {
		client := slack.New(org.SlackWebhookURL, true)
		go client.NotifyOnInvite(user, org, args.Email)
	}

	return nil, nil
}

func (*schemaResolver) AcceptUserInvite(ctx context.Context, args *struct {
	InviteToken string
}) (*orgInviteResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	// If the user is natively authenticated, require a verified email (if via SSO, we assume the SSO provider
	// has authenticated the user's email)
	if actor.Provider == "" {
		u, err := auth0.GetAuth0User(ctx)
		if err != nil {
			return nil, err
		}
		if !u.EmailVerified {
			// Don't add user to the org until email is verified. This will be a common failure mode,
			// so rather than return an error we return a response the client can handle.
			return &orgInviteResolver{emailVerified: false}, nil
		}
	}

	token, err := invite.ParseToken(args.InviteToken)
	if err != nil {
		return nil, err
	}
	org, err := store.Orgs.GetByID(ctx, token.OrgID)
	if err != nil {
		return nil, err
	}

	_, err = store.OrgMembers.Create(ctx, token.OrgID, actor.UID)
	if err != nil {
		return nil, err
	}

	if user, err := currentUser(ctx); err != nil {
		// errors swallowed because user is only needed for Slack notifications
		log15.Error("graphqlbackend.AcceptUserInvite: currentUser failed", "error", err)
	} else {
		client := slack.New(org.SlackWebhookURL, true)
		go client.NotifyOnAcceptedInvite(user, org)
	}

	return &orgInviteResolver{emailVerified: true}, nil
}
