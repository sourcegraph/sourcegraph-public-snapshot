package yaml_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/appliance/yaml"
)

func TestConvertsStringsToMultilineLiterals(t *testing.T) {
	doc := `regular_string: "a string"
multiline_string: "a\nmultiline\nstring"
number: 3
obj:
  foo: bar
`
	out, err := yaml.ConvertYAMLStringsToMultilineLiterals([]byte(doc))
	require.NoError(t, err)

	expected := `regular_string: "a string"
multiline_string: |-
  a
  multiline
  string
number: 3
obj:
  foo: bar
`

	require.Equal(t, expected, string(out))
}

func TestConvertsStringsContainingTrailingSpaceLinesToMultilineLiterals(t *testing.T) {
	doc := `regular_string: "a string"
multiline_string: "a\nmultiline   \nstring   "
number: 3
obj:
  foo: bar
`
	out, err := yaml.ConvertYAMLStringsToMultilineLiterals([]byte(doc))
	require.NoError(t, err)

	expected := `regular_string: "a string"
multiline_string: |-
  a
  multiline
  string
number: 3
obj:
  foo: bar
`

	require.Equal(t, expected, string(out))
}
