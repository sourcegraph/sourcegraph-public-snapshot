package config

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

// default json behaviour is to render nil slices as "null", so we manually
// set all nil slices in the struct to empty slice
func MarshalJSON(config IndexConfiguration) ([]byte, error) {
	nonNil := config
	if nonNil.IndexJobs == nil {
		nonNil.IndexJobs = []IndexJob{}
	}
	if nonNil.SharedSteps == nil {
		nonNil.SharedSteps = []DockerStep{}
	}
	for idx := range nonNil.SharedSteps {
		if nonNil.SharedSteps[idx].Commands == nil {
			nonNil.SharedSteps[idx].Commands = []string{}
		}
	}
	for idx := range nonNil.IndexJobs {
		if nonNil.IndexJobs[idx].IndexerArgs == nil {
			nonNil.IndexJobs[idx].IndexerArgs = []string{}
		}
		if nonNil.IndexJobs[idx].LocalSteps == nil {
			nonNil.IndexJobs[idx].LocalSteps = []string{}
		}
		if nonNil.IndexJobs[idx].Steps == nil {
			nonNil.IndexJobs[idx].Steps = []DockerStep{}
		}
		for stepIdx := range nonNil.IndexJobs[idx].Steps {
			if nonNil.IndexJobs[idx].Steps[stepIdx].Commands == nil {
				nonNil.IndexJobs[idx].Steps[stepIdx].Commands = []string{}
			}
		}
	}

	return json.MarshalIndent(nonNil, "", "    ")
}

func UnmarshalJSON(data []byte) (IndexConfiguration, error) {
	configuration := IndexConfiguration{}
	if err := jsonc.Unmarshal(string(data), &configuration); err != nil {
		return IndexConfiguration{}, fmt.Errorf("invalid JSON: %v", err)
	}
	return configuration, nil
}
