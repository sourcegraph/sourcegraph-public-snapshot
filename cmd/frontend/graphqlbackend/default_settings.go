package graphqlbackend

import (
	"context"
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var builtinExtensions = map[string]bool{
	"sourcegraph/clojure":    true,
	"sourcegraph/cobol":      true,
	"sourcegraph/cpp":        true,
	"sourcegraph/csharp":     true,
	"sourcegraph/cuda":       true,
	"sourcegraph/dart":       true,
	"sourcegraph/elixir":     true,
	"sourcegraph/erlang":     true,
	"sourcegraph/git-extras": true,
	"sourcegraph/go":         true,
	"sourcegraph/graphql":    true,
	"sourcegraph/groovy":     true,
	"sourcegraph/haskell":    true,
	"sourcegraph/java":       true,
	"sourcegraph/jsonnet":    true,
	"sourcegraph/kotlin":     true,
	"sourcegraph/lisp":       true,
	"sourcegraph/lua":        true,
	"sourcegraph/ocaml":      true,
	"sourcegraph/pascal":     true,
	"sourcegraph/perl":       true,
	"sourcegraph/php":        true,
	"sourcegraph/powershell": true,
	"sourcegraph/protobuf":   true,
	"sourcegraph/python":     true,
	"sourcegraph/r":          true,
	"sourcegraph/ruby":       true,
	"sourcegraph/rust":       true,
	"sourcegraph/scala":      true,
	"sourcegraph/shell":      true,
	"sourcegraph/swift":      true,
	"sourcegraph/tcl":        true,
	"sourcegraph/thrift":     true,
	"sourcegraph/typescript": true,
	"sourcegraph/verilog":    true,
	"sourcegraph/vhdl":       true,
}

const singletonDefaultSettingsGQLID = "DefaultSettings"

type defaultSettingsResolver struct {
	db    dbutil.DB
	gqlID string
}

func marshalDefaultSettingsGQLID(defaultSettingsID string) graphql.ID {
	return relay.MarshalID("DefaultSettings", defaultSettingsID)
}

func (r *defaultSettingsResolver) ID() graphql.ID { return marshalDefaultSettingsGQLID(r.gqlID) }

func (r *defaultSettingsResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	extensionIDs := []string{}
	for id := range builtinExtensions {
		extensionIDs = append(extensionIDs, id)
	}
	extensionIDs = ExtensionRegistry(r.db).FilterRemoteExtensions(extensionIDs)
	extensions := map[string]bool{}
	for _, id := range extensionIDs {
		extensions[id] = true
	}
	contents, err := json.Marshal(map[string]map[string]bool{"extensions": extensions})
	if err != nil {
		return nil, err
	}
	settings := &api.Settings{Subject: api.SettingsSubject{Default: true}, Contents: string(contents)}
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
