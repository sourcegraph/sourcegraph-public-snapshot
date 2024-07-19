package upgrades

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDetermineUpgradePolicy(t *testing.T) {
	tests := []struct {
		name             string
		currentVersion   string
		targetVersion    string
		expectedDowntime bool
		expectedError    bool
	}{
		{
			name:             "same version",
			currentVersion:   "5.0.0",
			targetVersion:    "5.0.0",
			expectedDowntime: false,
			expectedError:    false,
		},
		{
			name:             "standard version upgrade",
			currentVersion:   "3.43.0",
			targetVersion:    "3.44.0",
			expectedDowntime: false,
			expectedError:    false,
		},
		{
			name:             "major border case 1",
			currentVersion:   "3.43.0",
			targetVersion:    "4.0.0",
			expectedDowntime: false,
			expectedError:    false,
		},
		{
			name:             "major border case 2",
			currentVersion:   "4.5.0",
			targetVersion:    "5.0.0",
			expectedDowntime: false,
			expectedError:    false,
		},
		{
			name:             "major border mvu case",
			currentVersion:   "3.43.0",
			targetVersion:    "4.1.0",
			expectedDowntime: true,
			expectedError:    false,
		},
		{
			name:             "major border mvu case",
			currentVersion:   "5.0.0",
			targetVersion:    "4.1.0",
			expectedDowntime: false,
			expectedError:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			downtime, err := DetermineUpgradePolicy(test.currentVersion, test.targetVersion)
			if test.expectedError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !test.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if downtime != test.expectedDowntime {
				t.Errorf("expected downtime %v, got %v", test.expectedDowntime, downtime)
			}
		})
	}
}

func TestCheckConnection(t *testing.T) {
	tests := []struct {
		name       string
		schema     string
		connection bool
	}{
		{
			name:       "frontend single db connection",
			schema:     "frontend",
			connection: true,
		},
		{
			name:       "codeintel single db connection",
			schema:     "codeintel",
			connection: true,
		},
		{
			name:       "codeinsights single db connection",
			schema:     "codeinsights",
			connection: true,
		},
		{
			name:       "malformed dsn",
			schema:     "doombot",
			connection: false,
		},
	}

	// Setup mock db and attempt to connect using the given env vars
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// dbtest.NewDB(t)
			t.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")
			switch test.schema {
			case "frontend":
				t.Setenv("PGUSER", "sourcegraph")
			case "codeintel":
				t.Setenv("CODEINTEL_PGUSER", "sourcegraph")
			case "codeinsights":
				t.Setenv("CODEINSIGHTS_PGUSER", "sourcegraph")
			}

			dsns, err := getApplianceDSNs()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				t.FailNow()
			}

			t.Log("test dsn: ", dsns[test.schema])
			err = checkConnection(&observation.TestContext, test.schema, dsns[test.schema])
			if err != nil && test.connection {
				t.Errorf("unexpected error: %v", err)
				t.FailNow()
			}
		})
	}
}
