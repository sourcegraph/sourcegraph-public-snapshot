package graphqlbackend

import (
	"context"
	"encoding/json"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

var builtinExtensionIDs = []string{"sourcegraph/basic-code-intel"}

const singletonDefaultSettingsGQLID = "Default settings"

type defaultSettingsResolver struct {
	gqlID string
}

var singletonDefaultSettingsResolver = &defaultSettingsResolver{gqlID: singletonDefaultSettingsGQLID}

func marshalDefaultSettingsGQLID(defaultSettingsID string) graphql.ID {
	return relay.MarshalID("Default settings", defaultSettingsID)
}

func (r *defaultSettingsResolver) ID() graphql.ID { return marshalDefaultSettingsGQLID(r.gqlID) }

func (r *defaultSettingsResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	extensions := map[string]map[string]bool{"extensions": map[string]bool{}}
	for _, extensionID := range builtinExtensionIDs {
		extensions["extensions"][extensionID] = true
	}
	s, err := json.Marshal(extensions)
	if err != nil {
		return nil, err
	}
	settings := &api.Settings{Subject: api.SettingsSubject{Default: true}, Contents: string(s)}
	return &settingsResolver{&settingsSubject{defaultSettings: r}, settings, nil}, nil
}

func (r *defaultSettingsResolver) SettingsURL() string { return "/nonexistent" }

func (r *defaultSettingsResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return false, nil
}

func (r *defaultSettingsResolver) SettingsCascade() *settingsCascade {
	return &settingsCascade{subject: &settingsSubject{defaultSettings: r}}
}

func (r *defaultSettingsResolver) ConfigurationCascade() *settingsCascade { return r.SettingsCascade() }
