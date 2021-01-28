package config

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

func UnmarshalJSON(data []byte) (IndexConfiguration, error) {
	configuration := IndexConfiguration{}
	if err := jsonc.Unmarshal(string(data), &configuration); err != nil {
		return IndexConfiguration{}, fmt.Errorf("invalid JSON: %v", err)
	}
	return configuration, nil
}
