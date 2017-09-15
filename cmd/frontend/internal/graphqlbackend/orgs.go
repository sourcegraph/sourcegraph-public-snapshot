package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	jwt "github.com/dgrijalva/jwt-go"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

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
		member := &orgMemberResolver{sgMember}
		members = append(members, member)
	}
	return members, err
}

func (*schemaResolver) CreateOrg(ctx context.Context, args *struct {
	Name      string
	UserName  string
	UserEmail string
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
	_, err = store.OrgMembers.Create(ctx, int(newOrg.ID), actor.UID, args.UserName, args.UserEmail)
	if err != nil {
		return nil, err
	}

	return &orgResolver{org: newOrg}, nil
}

func (*schemaResolver) RemoveUserFromOrg(ctx context.Context, args *struct {
	UserID string
	OrgID  int32
}) (*EmptyResponse, error) {
	if isMember, err := store.Orgs.CurrentUserIsMember(ctx, store.OrgID(args.OrgID)); !isMember || err != nil {
		return nil, err
	}
	log15.Info("removing user from org", "user", args.UserID, "org", args.OrgID)
	return nil, store.OrgMembers.Remove(ctx, int(args.OrgID), args.UserID)
}

func (*schemaResolver) InviteUser(ctx context.Context, args *struct {
	UserEmail string
	OrgID     int32
}) (*EmptyResponse, error) {
	if isMember, err := store.Orgs.CurrentUserIsMember(ctx, store.OrgID(args.OrgID)); !isMember || err != nil {
		return nil, err
	}
	token, err := createOrgInviteToken(store.OrgID(args.OrgID))
	if err != nil {
		return nil, err
	}

	// TODO: send email
	log.Println(token)

	return nil, nil
}

func (*schemaResolver) AcceptUserInvite(ctx context.Context, args *struct {
	InviteToken string
	UserName    string
	UserEmail   string
}) (*orgMemberResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("no current user")
	}

	orgID, err := orgIDFromInviteToken(args.InviteToken)
	if err != nil {
		return nil, err
	}
	_, err = store.Orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	m, err := store.OrgMembers.Create(ctx, int(orgID), actor.UID, args.UserName, args.UserEmail)
	if err != nil {
		return nil, err
	}
	return &orgMemberResolver{member: m}, nil
}

func createOrgInviteToken(orgID store.OrgID) (string, error) {
	payload := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"orgID": int(orgID),
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
