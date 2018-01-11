package graphqlbackend

import (
	"context"
	"time"

	graphql "github.com/neelance/graphql-go"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

type settingsResolver struct {
	subject  *configurationSubject
	settings *sourcegraph.Settings
	user     *sourcegraph.User
}

func (o *settingsResolver) ID() int32 {
	return o.settings.ID
}

func (o *settingsResolver) Subject() *configurationSubject {
	return o.subject
}

func (o *settingsResolver) Configuration() *configurationResolver {
	return &configurationResolver{contents: o.settings.Contents}
}

func (o *settingsResolver) Contents() string { return o.settings.Contents }

func (o *settingsResolver) CreatedAt() string {
	return o.settings.CreatedAt.Format(time.RFC3339) // ISO
}

func (o *settingsResolver) Author(ctx context.Context) (*userResolver, error) {
	if o.user == nil {
		var err error
		o.user, err = db.Users.GetByID(ctx, o.settings.AuthorUserID)
		if err != nil {
			return nil, err
		}
	}
	return &userResolver{o.user}, nil
}

func (*schemaResolver) UpdateUserSettings(ctx context.Context, args *struct {
	LastKnownSettingsID *int32
	Contents            string
}) (*settingsResolver, error) {
	// 🚨 SECURITY: verify that the current user is authenticated.
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	settings, err := db.Settings.CreateIfUpToDate(ctx, sourcegraph.ConfigurationSubject{User: &user.ID}, args.LastKnownSettingsID, actor.FromContext(ctx).UID, args.Contents)
	if err != nil {
		return nil, err
	}
	return &settingsResolver{
		subject:  &configurationSubject{user: &userResolver{user: user}},
		settings: settings,
	}, nil
}

func (*schemaResolver) UpdateOrgSettings(ctx context.Context, args *struct {
	ID                  *graphql.ID
	OrgID               *graphql.ID // deprecated
	LastKnownSettingsID *int32
	Contents            string
}) (*settingsResolver, error) {
	orgID, err := unmarshalOrgGraphQLID(args.ID, args.OrgID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is a member of the org.
	if err := backend.CheckCurrentUserIsOrgMember(ctx, orgID); err != nil {
		return nil, err
	}

	org, err := db.Orgs.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	settings, err := db.Settings.CreateIfUpToDate(ctx, sourcegraph.ConfigurationSubject{Org: &orgID}, args.LastKnownSettingsID, actor.FromContext(ctx).UID, args.Contents)
	if err != nil {
		return nil, err
	}
	return &settingsResolver{
		subject:  &configurationSubject{org: &orgResolver{org}},
		settings: settings,
	}, nil
}

func currentSiteSettings(ctx context.Context) (*settingsResolver, error) {
	settings, err := db.Settings.GetLatest(ctx, sourcegraph.ConfigurationSubject{})
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, nil
	}
	return &settingsResolver{&configurationSubject{}, settings, nil}, nil
}

func (r *schemaResolver) CurrentSiteSettings(ctx context.Context) (*settingsResolver, error) {
	return currentSiteSettings(ctx)
}

func (*schemaResolver) UpdateSiteSettings(ctx context.Context, args *struct {
	LastKnownSettingsID *int32
	Contents            string
}) (*settingsResolver, error) {
	// 🚨 SECURITY: Only admins should be authorized to set global settings.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	settings, err := db.Settings.CreateIfUpToDate(ctx, sourcegraph.ConfigurationSubject{}, args.LastKnownSettingsID, actor.FromContext(ctx).UID, args.Contents)
	if err != nil {
		return nil, err
	}
	return &settingsResolver{
		subject:  &configurationSubject{},
		settings: settings,
	}, nil
}
