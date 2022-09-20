package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariableToGrafanaTemplateVar(t *testing.T) {
	t.Run("OptionsLabelValues", func(t *testing.T) {
		templateVar := (&ContainerVariable{
			OptionsLabelValues: ContainerVariableOptionsLabelValues{
				Query:     "metric",
				LabelName: "label",
			},
		}).toGrafanaTemplateVar()

		assert.Equal(t, templateVar.Query, "label_values(metric, label)")
	})
}

func TestVariableExampleValue(t *testing.T) {
	// Numbers get replaced with sentinel values - in getSentinelValue numbers are
	// replaced with the value 59m
	assert.Equal(t, "59m",
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
			OptionsLabelValues: ContainerVariableOptionsLabelValues{
				Query:         "bazbar",
				LabelName:     "asdf",
				ExampleOption: "foobar",
			},
		}).getSentinelValue())
}
