package config

import "testing"

func TestExtractIndexerName(t *testing.T) {
	tests := []struct {
		explanation string
		input       string
		expected    string
	}{
		{
			explanation: "no prefix",
			input:       "scip-go",
			expected:    "scip-go",
		},
		{
			explanation: "prefix",
			input:       "sourcegraph/scip-go",
			expected:    "scip-go",
		},
		{
			explanation: "prefix and suffix",
			input:       "sourcegraph/scip-go@sha256:...",
			expected:    "scip-go",
		},
		{
			explanation: "different name",
			input:       "myownscip-go",
			expected:    "myownscip-go",
		},
	}

	for _, test := range tests {
		t.Run(test.explanation, func(t *testing.T) {
			actual := extractIndexerName(test.input)
			if actual != test.expected {
				t.Errorf("unexpected indexer name. want=%q have=%q", test.expected, actual)
			}
		})
	}
}
