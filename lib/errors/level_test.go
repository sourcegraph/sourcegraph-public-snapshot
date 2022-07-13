package errors

import (
	"testing"
)

func TestClassifiedError(t *testing.T) {
	err := New("foo")

	tests := []struct {
		name  string
		err   error
		level Level
	}{
		{
			name:  "warn",
			err:   New("warn error"),
			level: LevelWarn,
		},
		{
			name:  "error",
			err:   err,
			level: LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just a quick test to make sure initialising this error works as intended.
			_ = NewClassifiedError(tt.err, tt.level)
		})
	}
}
