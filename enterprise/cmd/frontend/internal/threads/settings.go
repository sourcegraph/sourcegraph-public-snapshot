package threads

import (
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// TODO!(sqs)
type ThreadSettings struct {
	Delta *ThreadSettingsDelta `json:"delta"`
}

type ThreadSettingsDelta struct {
	Repository graphql.ID `json:"repository"`
	Base       string     `json:"base"`
	Head       string     `json:"head"`
}

func GetSettings(thread graphqlbackend.Thread) (*ThreadSettings, error) {
	var settings *ThreadSettings
	if err := json.Unmarshal([]byte(thread.Settings()), &settings); err != nil {
		return nil, err
	}
	return settings, nil
}
