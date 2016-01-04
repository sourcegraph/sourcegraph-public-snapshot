package worker

import (
	"testing"

	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
)

func TestMarshalConfigWithMatrix(t *testing.T) {
	config := droneyaml.Config{}

	tests := map[string]struct {
		axes []matrix.Axis
		want string
	}{
		"0x0": {
			axes: []matrix.Axis{{}},
			want: "{}\n",
		},
		"1x0": {
			axes: []matrix.Axis{
				map[string]string{"A": "a"},
			},
			want: `{}

matrix:
  A:
  - a
`,
		},
		"2x0": {
			axes: []matrix.Axis{
				map[string]string{"A": "a"},
				map[string]string{"A": "b"},
			},
			want: `{}

matrix:
  A:
  - a
  - b
`,
		},
		"2x1": {
			axes: []matrix.Axis{
				map[string]string{"A": "a", "C": "c"},
				map[string]string{"A": "b", "C": "c"},
			},
			want: `{}

matrix:
  A:
  - a
  - b
  C:
  - c
`,
		},
		"2x2": {
			axes: []matrix.Axis{
				map[string]string{"A": "a", "C": "c"},
				map[string]string{"A": "b", "C": "c"},
				map[string]string{"A": "a", "C": "d"},
				map[string]string{"A": "b", "C": "d"},
			},
			want: `{}

matrix:
  A:
  - a
  - b
  C:
  - c
  - d
`,
		},
		"2x2x3": {
			axes: []matrix.Axis{
				map[string]string{"A": "a", "C": "c", "E": "e"},
				map[string]string{"A": "b", "C": "c", "E": "e"},
				map[string]string{"A": "a", "C": "d", "E": "e"},
				map[string]string{"A": "b", "C": "d", "E": "e"},
				map[string]string{"A": "a", "C": "c", "E": "f"},
				map[string]string{"A": "b", "C": "c", "E": "f"},
				map[string]string{"A": "a", "C": "d", "E": "f"},
				map[string]string{"A": "b", "C": "d", "E": "f"},
				map[string]string{"A": "a", "C": "c", "E": "g"},
				map[string]string{"A": "b", "C": "c", "E": "g"},
				map[string]string{"A": "a", "C": "d", "E": "g"},
				map[string]string{"A": "b", "C": "d", "E": "g"},
			},
			want: `{}

matrix:
  A:
  - a
  - b
  C:
  - c
  - d
  E:
  - e
  - f
  - g
`,
		},
	}

	for label, test := range tests {
		yaml, err := marshalConfigWithMatrix(config, test.axes)
		if err != nil {
			t.Errorf("%s: %s", label, err)
			continue
		}
		if string(yaml) != test.want {
			t.Errorf("%s: got %q, want %q", label, yaml, test.want)
			continue
		}
	}
}
