package plan

import (
	"sort"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"gopkg.in/yaml.v2"
)

func parseMatrix(yaml string) ([]matrix.Axis, error) {
	axes, err := matrix.Parse(yaml)
	if err != nil {
		return nil, err
	}
	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}
	return axes, nil
}

func calcMatrix(m matrix.Matrix) []matrix.Axis {
	axes := matrix.Calc(m)
	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}
	return axes
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
