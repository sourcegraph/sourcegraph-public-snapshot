package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariableExampleValue(t *testing.T) {
	// Numbers get replaced with sentinel values - in getSentinelValue numbers are
	// replaced with the value 1234
	assert.Equal(t, "1234m",
		(&ContainerVariable{
			Options: ContainerVariableOptions{
				Options: []string{
					"1m",
					"5m",
					"60m",
				},
			},
		}).getSentinelValue())

	// Strings do not
	assert.Equal(t, "foobar",
		(&ContainerVariable{
			OptionsQuery: ContainerVariableOptionsQuery{
				Query:         "bazbar",
				ExampleOption: "foobar",
			},
		}).getSentinelValue())
}
