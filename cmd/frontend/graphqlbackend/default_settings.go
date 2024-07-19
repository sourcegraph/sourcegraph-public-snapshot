package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

const singletonDefaultSettingsGQLID = "DefaultSettings"

func newDefaultSettingsResolver(db database.DB) *defaultSettingsResolver {
	return &defaultSettingsResolver{
		db:    db,
		gqlID: singletonDefaultSettingsGQLID,
	}
}

type defaultSettingsResolver struct {
	db    database.DB
	gqlID string
}

func marshalDefaultSettingsGQLID(defaultSettingsID string) graphql.ID {
	return relay.MarshalID("DefaultSettings", defaultSettingsID)
}

func (r *defaultSettingsResolver) ID() graphql.ID { return marshalDefaultSettingsGQLID(r.gqlID) }

func (r *defaultSettingsResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	// 🚨 SECURITY: Check that the viewer can access these settings.
	subject, err := settingsSubjectForNodeAndCheckAccess(ctx, r)
	if err != nil {
		return nil, err
	}

	settings := &api.Settings{
		Subject:  api.SettingsSubject{Default: true},
		Contents: `{"experimentalFeatures": {}}`,
	}
	return &settingsResolver{db: r.db, subject: subject, settings: settings}, nil
}

func (r *defaultSettingsResolver) SettingsURL() *string { return nil }

func (r *defaultSettingsResolver) ViewerCanAdminister(_ context.Context) (bool, error) {
	return false, nil
}

func (r *defaultSettingsResolver) SettingsCascade(ctx context.Context) (*settingsCascade, error) {
	// 🚨 SECURITY: Check that the viewer can access these settings.
	subject, err := settingsSubjectForNodeAndCheckAccess(ctx, r)
	if err != nil {
		return nil, err
	}
	return &settingsCascade{db: r.db, subject: subject}, nil
}

func (r *defaultSettingsResolver) ConfigurationCascade(ctx context.Context) (*settingsCascade, error) {
	return r.SettingsCascade(ctx)
}
