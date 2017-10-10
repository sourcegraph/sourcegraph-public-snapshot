package graphqlbackend

import (
	"context"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type orgSettingsResolver struct {
	org      *sourcegraph.Org
	settings *sourcegraph.OrgSettings
	user     *sourcegraph.User
}

func (o *orgSettingsResolver) ID() int32 {
	return o.settings.ID
}

func (o *orgSettingsResolver) Contents() string {
	return o.settings.Contents
}

func (o *orgSettingsResolver) CreatedAt() string {
	return o.settings.CreatedAt.Format(time.RFC3339) // ISO
}

func (o *orgSettingsResolver) Author(ctx context.Context) (*userResolver, error) {
	if o.user == nil {
		var err error
		o.user, err = store.Users.GetByAuth0ID(o.settings.AuthorAuth0ID)
		if err != nil {
			return nil, err
		}
	}
	return &userResolver{o.user, nil}, nil
}

func (*schemaResolver) UpdateOrgSettings(ctx context.Context, args *struct {
	OrgID    int32
	Contents string
}) (*orgSettingsResolver, error) {
	// 🚨 SECURITY: verify that the current user is in the org.
	actor := actor.FromContext(ctx)
	_, err := store.OrgMembers.GetByOrgIDAndUserID(ctx, args.OrgID, actor.UID)
	if err != nil {
		return nil, err
	}

	org, err := store.Orgs.GetByID(ctx, args.OrgID)
	if err != nil {
		return nil, err
	}

	setting, err := store.OrgSettings.Create(ctx, args.OrgID, actor.UID, args.Contents)
	if err != nil {
		return nil, err
	}
	return &orgSettingsResolver{org, setting, nil}, nil
}
