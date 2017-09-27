package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/conf"
	appconf "sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"

	"gopkg.in/inconshreveable/log15.v2"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mattbaird/gochimp"

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

func (o *orgResolver) Members(ctx context.Context) ([]*orgMemberResolver, error) {
	sgMembers, err := store.OrgMembers.GetByOrgID(ctx, o.org.ID)
	if err != nil {
		return nil, err
	}

	members := []*orgMemberResolver{}
	for _, sgMember := range sgMembers {
		member := &orgMemberResolver{o.org, sgMember}
		members = append(members, member)
	}
	return members, nil
}

func (o *orgResolver) Threads(ctx context.Context, args *struct {
	File  *string
	Limit *int32
}) ([]*threadResolver, error) {
	limit := int32(1000)
	if args.Limit != nil && *args.Limit < limit {
		limit = *args.Limit
	}
	threads, err := store.Threads.GetByOrg(ctx, o.org.ID, args.File, limit)
	if err != nil {
		return nil, err
	}
	threadResolvers := []*threadResolver{}
	for _, thread := range threads {
		threadResolvers = append(threadResolvers, &threadResolver{o.org, nil, thread})
	}
	return threadResolvers, nil
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
	Username    string
	Email       string
	DisplayName string
	AvatarURL   *string
}) (*orgResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	newOrg, err := store.Orgs.Create(ctx, args.Name)
	if err != nil {
		return nil, err
	}

	// Add the current user as the first member of the new org.
	_, err = store.OrgMembers.Create(ctx, newOrg.ID, actor.UID, args.Username, args.Email, args.DisplayName, args.AvatarURL)
	if err != nil {
		return nil, err
	}
	return &orgResolver{org: newOrg}, nil
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

	// Don't invite the user if they are already a member.
	_, err = store.OrgMembers.GetByOrgAndEmail(ctx, args.OrgID, args.Email)
	if err == nil {
		return nil, fmt.Errorf("user %s is already a member of org %d", args.Email, args.OrgID)
	}
	if _, ok := err.(store.ErrOrgMemberNotFound); !ok {
		return nil, err
	}

	org, err := localstore.Orgs.GetByID(ctx, args.OrgID)
	if err != nil {
		return nil, err
	}

	token, err := createOrgInviteToken(args.OrgID)
	if err != nil {
		return nil, err
	}

	sendInviteEmail(args.Email, orgMember.DisplayName, org.Name, token)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (*schemaResolver) AcceptUserInvite(ctx context.Context, args *struct {
	InviteToken string
	Username    string
	Email       string
	DisplayName string
	AvatarURL   *string
}) (*orgMemberResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	orgID, err := orgIDFromInviteToken(args.InviteToken)
	if err != nil {
		return nil, err
	}
	org, err := store.Orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	member, err := store.OrgMembers.Create(ctx, orgID, actor.UID, args.Username, args.Email, args.DisplayName, args.AvatarURL)
	if err != nil {
		return nil, err
	}
	return &orgMemberResolver{org, member}, nil
}

func createOrgInviteToken(orgID int32) (string, error) {
	payload := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"orgID": orgID,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	return payload.SignedString(conf.AppSecretKey)
}

func orgIDFromInviteToken(tokenString string) (int32, error) {
	payload, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("error parsing org invite: unexpected signing method %v", token.Header["alg"])
		}
		return conf.AppSecretKey, nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := payload.Claims.(jwt.MapClaims)
	if !ok && !payload.Valid {
		return 0, errors.New("error parsing org invite: invalid token")
	}
	id, ok := claims["orgID"].(float64)
	if !ok {
		return 0, errors.New("error parsing org invite: invalid type for field orgID")
	}
	return int32(id), nil
}

func sendInviteEmail(inviteEmail, fromName, orgName, token string) {
	config := &notif.EmailConfig{
		Template:  "invite-user",
		FromName:  fromName,
		FromEmail: "noreply@sourcegraph.com",
		ToEmail:   inviteEmail,
		Subject:   fmt.Sprintf("%s has invited you to join %s on Sourcegraph", fromName, orgName),
	}

	inviteURL := appconf.AppURL.String() + "/accept-invite?token=" + token
	notif.SendMandrillTemplate(config, []gochimp.Var{}, []gochimp.Var{
		gochimp.Var{Name: "INVITE_URL", Content: inviteURL},
		gochimp.Var{Name: "ORG", Content: orgName},
		gochimp.Var{Name: "FROM_USER", Content: fromName},
	})
}
