package conf

import (
	"context"
	"testing"
)

func TestIsGRPCEnabled(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name: "enabled",

			envValue: "true",
			expected: true,
		},
		{
			name: "disabled",

			envValue: "false",
			expected: false,
		},
		{
			name: "empty env var - default true",

			envValue: "",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(envGRPCEnabled, test.envValue)
			actual := IsGRPCEnabled(context.Background())

			if actual != test.expected {
				t.Errorf("expected %v but got %v", test.expected, actual)
			}
		})
	}

}
