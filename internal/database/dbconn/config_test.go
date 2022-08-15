package dbconn

import (
	"testing"

	"github.com/sourcegraph/log/logtest"
)

func TestBuildConfig(t *testing.T) {
	logger := logtest.Scoped(t)
	tests := []struct {
		name                    string
		dataSource              string
		expectedApplicationName string
		fails                   bool
	}{
		{
			name:                    "empty dataSource",
			dataSource:              "",
			expectedApplicationName: defaultApplicationName,
			fails:                   false,
		}, {
			name:                    "connection string",
			dataSource:              "dbname=sourcegraph host=localhost sslmode=verify-full user=sourcegraph",
			expectedApplicationName: defaultApplicationName,
			fails:                   false,
		}, {
			name:                    "connection string with application name",
			dataSource:              "dbname=sourcegraph host=localhost sslmode=verify-full user=sourcegraph application_name=foo",
			expectedApplicationName: "foo",
			fails:                   false,
		}, {
			name:                    "postgres URL",
			dataSource:              "postgres://sourcegraph@localhost/sourcegraph?sslmode=verify-full",
			expectedApplicationName: defaultApplicationName,
			fails:                   false,
		}, {
			name:                    "postgres URL with fallback",
			dataSource:              "postgres://sourcegraph@localhost/sourcegraph?sslmode=verify-full&application_name=foo",
			expectedApplicationName: "foo",
			fails:                   false,
		}, {
			name:       "invalid URL",
			dataSource: "invalid string",
			fails:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := buildConfig(logger, tt.dataSource, "")
			if tt.fails {
				if err == nil {
					t.Fatal("error expected")
				}

				return
			}

			fb, ok := cfg.RuntimeParams["application_name"]
			if !ok || fb != tt.expectedApplicationName {
				t.Fatalf("wrong application_name: got %q want %q", fb, tt.expectedApplicationName)
			}
		})
	}
}
