package config

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestValidateConfig(t *testing.T) {
	t.Run("Frontend URL", func(t *testing.T) {
		tests := []struct {
			name        string
			frontendURL string
			expectedErr error
		}{
			{
				name:        "Valid URL",
				frontendURL: "https://sourcegraph.example.com",
				expectedErr: nil,
			},
			{
				name:        "Missing scheme",
				frontendURL: "sourcegraph.example.com",
				expectedErr: errors.New("EXECUTOR_FRONTEND_URL must be in the format scheme://host (and optionally :port)"),
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				conf := Config{
					FrontendURL:    test.frontendURL,
					QueueName:      "batches",
					UseFirecracker: false,
				}

				err := conf.Validate()
				if !errors.Is(err, test.expectedErr) {
					t.Errorf("Unexpected error returned: expected '%v', got '%v'", test.expectedErr, err)
				}
			})
		}
	})
}
