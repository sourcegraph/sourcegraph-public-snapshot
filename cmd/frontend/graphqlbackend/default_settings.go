package graphqlbackend

import (
	"context"
	"encoding/json"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

var builtinExtensions = map[string]bool{
	"sourcegraph/typescript": true,
	"sourcegraph/python":     true,
	"sourcegraph/java":       true,
	"sourcegraph/go":         true,
	"sourcegraph/cpp":        true,
	"sourcegraph/ruby":       true,
	"sourcegraph/php":        true,
	"sourcegraph/csharp":     true,
	"sourcegraph/shell":      true,
	"sourcegraph/scala":      true,
	"sourcegraph/kotlin":     true,
	"sourcegraph/r":          true,
	"sourcegraph/perl":       true,
}

const singletonDefaultSettingsGQLID = "DefaultSettings"

type defaultSettingsResolver struct {
	gqlID string
}

var singletonDefaultSettingsResolver = &defaultSettingsResolver{gqlID: singletonDefaultSettingsGQLID}

func marshalDefaultSettingsGQLID(defaultSettingsID string) graphql.ID {
	return relay.MarshalID("DefaultSettings", defaultSettingsID)
}

func (r *defaultSettingsResolver) ID() graphql.ID { return marshalDefaultSettingsGQLID(r.gqlID) }

func (r *defaultSettingsResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	extensionIDs := []string{}
	for id := range builtinExtensions {
		extensionIDs = append(extensionIDs, id)
	}
	extensionIDs = ExtensionRegistry.FilterRemoteExtensions(extensionIDs)
	extensions := map[string]bool{}
	for _, id := range extensionIDs {
		extensions[id] = true
	}
	contents, err := json.Marshal(map[string]map[string]bool{"extensions": extensions})
	if err != nil {
		return nil, err
	}
	settings := &api.Settings{Subject: api.SettingsSubject{Default: true}, Contents: string(contents)}
	return &settingsResolver{&settingsSubject{defaultSettings: r}, settings, nil}, nil
}

func (r *defaultSettingsResolver) SettingsURL() *string { return nil }

func (r *defaultSettingsResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return false, nil
}

func (r *defaultSettingsResolver) SettingsCascade() *settingsCascade {
	return &settingsCascade{subject: &settingsSubject{defaultSettings: r}}
}

func (r *defaultSettingsResolver) ConfigurationCascade() *settingsCascade { return r.SettingsCascade() }
