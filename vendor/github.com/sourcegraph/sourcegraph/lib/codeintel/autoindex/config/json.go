package config

import (
	"encoding/json"
	"strings"

	"github.com/sourcegraph/jsonx"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// default json behaviour is to render nil slices as "null", so we manually
// set all nil slices in the struct to empty slice
func MarshalJSON(config AutoIndexJobSpecList) ([]byte, error) {
	nonNil := config
	if nonNil.JobSpecs == nil {
		nonNil.JobSpecs = []AutoIndexJobSpec{}
	}
	for idx := range nonNil.JobSpecs {
		if nonNil.JobSpecs[idx].IndexerArgs == nil {
			nonNil.JobSpecs[idx].IndexerArgs = []string{}
		}
		if nonNil.JobSpecs[idx].LocalSteps == nil {
			nonNil.JobSpecs[idx].LocalSteps = []string{}
		}
		if nonNil.JobSpecs[idx].Steps == nil {
			nonNil.JobSpecs[idx].Steps = []DockerStep{}
		}
		for stepIdx := range nonNil.JobSpecs[idx].Steps {
			if nonNil.JobSpecs[idx].Steps[stepIdx].Commands == nil {
				nonNil.JobSpecs[idx].Steps[stepIdx].Commands = []string{}
			}
		}
	}

	return json.MarshalIndent(nonNil, "", "    ")
}

func UnmarshalJSON(data []byte) (AutoIndexJobSpecList, error) {
	configuration := AutoIndexJobSpecList{}
	if err := jsonUnmarshal(string(data), &configuration); err != nil {
		return AutoIndexJobSpecList{}, errors.Errorf("invalid JSON: %v", err)
	}
	return configuration, nil
}

// jsonUnmarshal unmarshals the JSON using a fault-tolerant parser that allows comments
// and trailing commas. If any unrecoverable faults are found, an error is returned.
func jsonUnmarshal(text string, v any) error {
	data, errs := jsonx.Parse(text, jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(errs) > 0 {
		return errors.Errorf("failed to parse JSON: %v", errs)
	}
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return json.Unmarshal(data, v)
}
