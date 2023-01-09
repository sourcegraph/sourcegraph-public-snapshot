package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVariableToGrafanaTemplateVar(t *testing.T) {
	t.Run("OptionsLabelValues", func(t *testing.T) {
		templateVar, err := (&ContainerVariable{
			OptionsLabelValues: ContainerVariableOptionsLabelValues{
				Query:     "metric",
				LabelName: "label",
			},
		}).toGrafanaTemplateVar(nil)

		assert.Nil(t, err)
		assert.Equal(t, templateVar.Query, "label_values(metric, label)")
	})
}

func TestVariableExampleValue(t *testing.T) {
	// Numbers get replaced with sentinel values - in getSentinelValue numbers are
	// replaced with the value 60-len(name)
	assert.Equal(t, "56m",
		(&ContainerVariable{
			Name: "name",
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

func TestVariableApplier(t *testing.T) {
	t.Run("with intervals", func(t *testing.T) {
		vars := newVariableApplier([]ContainerVariable{
			{
				Name: "foo",
				Options: ContainerVariableOptions{
					Options: []string{"1m"},
				},
			},
			{
				Name: "bar",
				OptionsLabelValues: ContainerVariableOptionsLabelValues{
					ExampleOption: "hello-world",
				},
			},
		})

		var expression = `metric{bar="$bar"}[$foo]`

		applied := vars.ApplySentinelValues(expression)
		assert.Equal(t, `metric{bar="$bar"}[57m]`, applied) // sentinel value is 60-len(name)

		reverted := vars.RevertDefaults(expression, applied)
		assert.Equal(t, `metric{bar="$bar"}[$foo]`, reverted)
	})

	t.Run("without intervals", func(t *testing.T) {
		vars := newVariableApplierWith([]ContainerVariable{
			{
				Name: "foo",
				Options: ContainerVariableOptions{
					Options: []string{"1m"},
				},
			},
			{
				Name: "bar",
				OptionsLabelValues: ContainerVariableOptionsLabelValues{
					ExampleOption: "hello-world",
				},
			},
		}, false)

		var expression = `metric{bar="$bar"}[$foo]`

		applied := vars.ApplySentinelValues(expression)
		// no replacement for intervals
		assert.Equal(t, `metric{bar="$bar"}[$foo]`, applied)

		reverted := vars.RevertDefaults(expression, applied)
		assert.Equal(t, `metric{bar="$bar"}[$foo]`, reverted)
	})
}
