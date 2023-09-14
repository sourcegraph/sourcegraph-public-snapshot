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

func (r *defaultSettingsResolver) LatestSettings(_ context.Context) (*settingsResolver, error) {
	settings := &api.Settings{
		Subject:  api.SettingsSubject{Default: true},
		Contents: `{"experimentalFeatures": {}}`,
	}
	return &settingsResolver{r.db, &settingsSubjectResolver{defaultSettings: r}, settings, nil}, nil
}

func (r *defaultSettingsResolver) SettingsURL() *string { return nil }

func (r *defaultSettingsResolver) ViewerCanAdminister(_ context.Context) (bool, error) {
	return false, nil
}

func (r *defaultSettingsResolver) SettingsCascade() *settingsCascade {
	return &settingsCascade{db: r.db, subject: &settingsSubjectResolver{defaultSettings: r}}
}

func (r *defaultSettingsResolver) ConfigurationCascade() *settingsCascade { return r.SettingsCascade() }
