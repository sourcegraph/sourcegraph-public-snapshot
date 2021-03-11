package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

func UnmarshalYAML(data []byte) (IndexConfiguration, error) {
	configuration := IndexConfiguration{}
	if err := yaml.Unmarshal(data, &configuration); err != nil {
		return IndexConfiguration{}, fmt.Errorf("invalid YAML: %v", err)
	}
	return configuration, nil
}
