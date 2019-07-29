package changesets

import (
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// TODO!(sqs)
type ChangesetSettings struct {
	Delta *ChangesetSettingsDelta `json:"delta"`
}

type ChangesetSettingsDelta struct {
	Repository graphql.ID `json:"repository"`
	Base       string     `json:"base"`
	Head       string     `json:"head"`
}

func GetSettings(changeset graphqlbackend.Changeset) (*ChangesetSettings, error) {
	var settings *ChangesetSettings
	if err := json.Unmarshal([]byte(changeset.Settings()), &settings); err != nil {
		return nil, err
	}
	return settings, nil
}
