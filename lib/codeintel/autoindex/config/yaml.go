package config

import (
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func UnmarshalYAML(data []byte) (IndexConfiguration, error) {
	configuration := IndexConfiguration{}
	if err := yaml.Unmarshal(data, &configuration); err != nil {
		return IndexConfiguration{}, errors.Errorf("invalid YAML: %v", err)
	}

	return configuration, nil
}
