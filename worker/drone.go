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
	dronerunner.DefaultCloner = "sourcegraph/drone-git"
	droneparser.DefaultCloner = "sourcegraph/drone-git"
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
