package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

const singletonDefaultSettingsGQLID = "DefaultSettings"

type defaultSettingsResolver struct {
	db    database.DB
	gqlID string
}

func marshalDefaultSettingsGQLID(defaultSettingsID string) graphql.ID {
	return relay.MarshalID("DefaultSettings", defaultSettingsID)
}

func (r *defaultSettingsResolver) ID() graphql.ID { return marshalDefaultSettingsGQLID(r.gqlID) }

func (r *defaultSettingsResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	settings := &api.Settings{Subject: api.SettingsSubject{Default: true}, Contents: "{}"}
	return &settingsResolver{r.db, &settingsSubject{defaultSettings: r}, settings, nil}, nil
}

func (r *defaultSettingsResolver) SettingsURL() *string { return nil }

func (r *defaultSettingsResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return false, nil
}

func (r *defaultSettingsResolver) SettingsCascade() *settingsCascade {
	return &settingsCascade{db: r.db, subject: &settingsSubject{defaultSettings: r}}
}

func (r *defaultSettingsResolver) ConfigurationCascade() *settingsCascade { return r.SettingsCascade() }
