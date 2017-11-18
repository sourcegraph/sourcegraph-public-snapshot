package graphqlbackend

import (
	"context"
	"errors"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/highlight"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

type settingsSubject struct {
	org *orgResolver
}

func (s *settingsSubject) ToOrg() (*orgResolver, bool) { return s.org, s.org != nil }

func (s *settingsSubject) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	switch {
	case s.org != nil:
		return s.org.LatestSettings(ctx)
	}
	panic("no settings subject type")
}

type settingsResolver struct {
	subject  *settingsSubject
	settings *sourcegraph.Settings
	user     *sourcegraph.User
}

func (o *settingsResolver) ID() int32 {
	return o.settings.ID
}

func (o *settingsResolver) Subject() *settingsSubject {
	return o.subject
}

func (o *settingsResolver) Contents() string {
	return o.settings.Contents
}

func (o *settingsResolver) Highlighted(ctx context.Context) (string, error) {
	html, aborted, err := highlight.Code(ctx, o.Contents(), "json", false)
	if err != nil {
		return "", err
	}
	if aborted {
		// Settings should be small enough so the syntax highlighting
		// completes before the automatic timeout. If it doesn't, something
		// seriously wrong has happened.
		return "", errors.New("settings syntax highlighting aborted")
	}

	return string(html), nil
}

func (o *settingsResolver) CreatedAt() string {
	return o.settings.CreatedAt.Format(time.RFC3339) // ISO
}

func (o *settingsResolver) Author(ctx context.Context) (*userResolver, error) {
	if o.user == nil {
		var err error
		o.user, err = store.Users.GetByAuth0ID(ctx, o.settings.AuthorAuth0ID)
		if err != nil {
			return nil, err
		}
	}
	return &userResolver{o.user, nil}, nil
}

func (*schemaResolver) UpdateOrgSettings(ctx context.Context, args *struct {
	OrgID               int32
	LastKnownSettingsID *int32
	Contents            string
}) (*settingsResolver, error) {
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

	settings, err := store.Settings.CreateIfUpToDate(ctx, args.OrgID, args.LastKnownSettingsID, actor.UID, args.Contents)
	if err != nil {
		return nil, err
	}
	return &settingsResolver{
		subject:  &settingsSubject{org: &orgResolver{org}},
		settings: settings,
	}, nil
}
