package internalerrs

import (
	"testing"
)

func TestIsLoggingEnabled(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		want   bool
	}{
		{
			name: "env var not set",
			want: true,
		},
		{
			name:   "env var set to true",
			envVar: "true",
			want:   false,
		},
		{
			name:   "env var set to false",
			envVar: "false",
			want:   true,
		},
		{
			name:   "env var set to invalid value",
			envVar: "foo",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv(envDisableGRPCInternalErrorLogging, tt.envVar)
			}
			if got := isLoggingEnabled(); got != tt.want {
				t.Errorf("isLoggingEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
