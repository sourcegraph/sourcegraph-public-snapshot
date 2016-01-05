package worker

import (
	"sort"

	"gopkg.in/yaml.v2"

	droneparser "github.com/drone/drone-exec/parser"
	dronerunner "github.com/drone/drone-exec/runner"
	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
)

// Drone CI-related customizations and hacks.

func init() {
	// This image adds multiple netrc support and the "complete" param.
	//
	// We can switch back to the upstream plugin/drone-git when these
	// PRs are merged:
	// https://github.com/drone-plugins/drone-git/pull/14 and the
	// not-yet-submitted PR based on the github.com/sqs/drone-git
	// multiple-netrc-entries branch.
	dronerunner.DefaultCloner = "sourcegraph/drone-git@sha256:4f8133161cdbfe409b107dc73274635151d8f4bf3c639ff55884a5aa6d765311"
	droneparser.DefaultCloner = dronerunner.DefaultCloner
}

func marshalConfigWithMatrix(c droneyaml.Config, axes []matrix.Axis) ([]byte, error) {
	yamlBytes, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	hasVal := func(vals []string, val string) bool {
		for _, v := range vals {
			if v == val {
				return true
			}
		}
		return false
	}
	m := matrix.Matrix{}
	for _, axis := range axes {
		for k, v := range axis {
			if !hasVal(m[k], v) {
				m[k] = append(m[k], v)
			}
		}
	}

	if len(m) > 0 {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var ym yaml.MapSlice
		for _, k := range keys {
			ym = append(ym, yaml.MapItem{Key: k, Value: m[k]})
		}

		matrixBytes, err := yaml.Marshal(map[string]interface{}{"matrix": ym})
		if err != nil {
			return nil, err
		}
		yamlBytes = append(yamlBytes, '\n')
		yamlBytes = append(yamlBytes, matrixBytes...)
	}

	return yamlBytes, nil
}
