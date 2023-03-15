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
			input:       "lsif-go",
			expected:    "lsif-go",
		},
		{
			explanation: "prefix",
			input:       "sourcegraph/lsif-go",
			expected:    "lsif-go",
		},
		{
			explanation: "prefix and suffix",
			input:       "sourcegraph/lsif-go@sha256:...",
			expected:    "lsif-go",
		},
		{
			explanation: "different name",
			input:       "myownlsif-go",
			expected:    "myownlsif-go",
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
