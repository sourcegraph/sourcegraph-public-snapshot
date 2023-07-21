package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCheckSourcegraphVersion(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		constraint     string
		minDate        string
		expected       bool
		expectedErr    error
	}{
		{
			name:           "Version matches constraint",
			currentVersion: "3.12.6",
			constraint:     ">= 3.12.6",
			minDate:        "2020-01-19",
			expected:       true,
		},
		{
			name:           "Release candidate version matches constraint",
			currentVersion: "3.12.6-rc.1",
			constraint:     ">= 3.12.6-0",
			minDate:        "2020-01-19",
			expected:       true,
		},
		{
			name:           "Newer release candidate version matches constraint",
			currentVersion: "3.12.6-rc.3",
			constraint:     ">= 3.10.6-0",
			minDate:        "2020-01-19",
			expected:       true,
		},
		{
			name:           "Version does not match constraint",
			currentVersion: "3.12.6",
			constraint:     ">= 3.13",
			minDate:        "2020-01-19",
			expected:       false,
		},
		{
			name:           "Constraint without patch version",
			currentVersion: "3.13.0",
			constraint:     ">= 3.13",
			minDate:        "2020-01-19",
			expected:       true,
		},
		{
			name:           "Dev version",
			currentVersion: "dev",
			constraint:     ">= 3.13",
			minDate:        "2020-01-19",
			expected:       true,
		},
		{
			name:           "Newer dev version",
			currentVersion: "0.0.0+dev",
			constraint:     ">= 3.13",
			minDate:        "2020-01-19",
			expected:       true,
		},
		{
			name:           "Seven character abbreviated hash",
			currentVersion: "54959_2020-01-29_9258595",
			minDate:        "2020-01-19",
			constraint:     ">= 999.13",
			expected:       true,
		},
		{
			name:           "Seven character abbreviated hash too old",
			currentVersion: "54959_2020-01-29_9258595",
			minDate:        "2020-01-30",
			constraint:     ">= 999.13",
			expected:       false,
		},
		{
			name:           "Seven character abbreviated hash matches date",
			currentVersion: "54959_2020-01-29_9258595",
			minDate:        "2020-01-29",
			constraint:     ">= 0.0",
			expected:       true,
		},
		{
			name:           "Twelve character abbreviated hash",
			currentVersion: "54959_2020-01-29_925859585436",
			minDate:        "2020-01-19",
			constraint:     ">= 999.13",
			expected:       true,
		},
		{
			name:           "Twelve character abbreviated hash too old",
			currentVersion: "54959_2020-01-29_925859585436",
			minDate:        "2020-01-30",
			constraint:     ">= 999.13",
			expected:       false,
		},
		{
			name:           "Twelve character abbreviated hash matches date",
			currentVersion: "54959_2020-01-29_925859585436",
			minDate:        "2020-01-29",
			constraint:     ">= 0.0",
			expected:       true,
		},
		{
			name:           "Twelve character abbreviated hash with tag",
			currentVersion: "54959_2020-01-29_4.4-925859585436",
			minDate:        "2020-01-19",
			constraint:     ">= 999.13",
			expected:       true,
		},
		{
			name:           "Twelve character abbreviated hash with tag too old and does not match constraint",
			currentVersion: "54959_2020-01-29_4.4-925859585436",
			minDate:        "2020-01-30",
			constraint:     ">= 999.13",
			expected:       false,
		},
		{
			name:           "Twelve character abbreviated hash with tag matches date",
			currentVersion: "54959_2020-01-29_4.4-925859585436",
			minDate:        "2020-01-29",
			constraint:     ">= 0.0",
			expected:       true,
		},
		{
			name:           "Forty character hash",
			currentVersion: "54959_2020-01-29_7db7d396346284fd0f8f79f130f38b16fb1d3d70",
			minDate:        "2020-01-29",
			constraint:     ">= 0.0",
			expected:       true,
		},
		{
			name:           "Daily release build",
			currentVersion: "5.1_231128_2023-06-27_5.0-7ac9ba347103",
			minDate:        "2020-01-29",
			constraint:     ">= 4.4",
			expected:       true,
		},
		{
			name:           "Invalid semantic version",
			currentVersion: "\n1.2",
			minDate:        "2020-01-29",
			constraint:     ">= 0.0",
			expected:       false,
			expectedErr:    errors.New("Invalid Semantic Version"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := CheckSourcegraphVersion(test.currentVersion, test.constraint, test.minDate)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}
