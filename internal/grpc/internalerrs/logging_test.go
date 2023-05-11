package internalerrs

import (
	"testing"
)

func TestGetEnvWithDefaultBool(t *testing.T) {
	testCases := []struct {
		name           string
		envValue       string
		defaultValue   bool
		expectedResult bool
	}{
		{
			name: "Valid env value true",

			envValue:       "true",
			defaultValue:   false,
			expectedResult: true,
		},
		{
			name: "Valid env value false",

			envValue:       "false",
			defaultValue:   true,
			expectedResult: false,
		},
		{
			name: "Invalid env value, default true",

			envValue:       "invalid",
			defaultValue:   true,
			expectedResult: true,
		},
		{
			name: "Invalid env value, default false",

			envValue:       "invalid",
			defaultValue:   false,
			expectedResult: false,
		},
		{
			name: "Empty env value, default true",

			envValue:       "",
			defaultValue:   true,
			expectedResult: true,
		},
		{
			name: "Empty env value, default false",

			envValue:       "",
			defaultValue:   false,
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			envVar := "EXAMPLE_ENV_VAR"
			t.Setenv(envVar, tc.envValue)

			result := getEnvWithDefaultBool(envVar, tc.defaultValue)
			if result != tc.expectedResult {
				t.Errorf("Expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}
