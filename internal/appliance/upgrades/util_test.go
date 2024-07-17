package upgrades

import (
	"os"
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
		connection bool
		envvars    map[string]string
	}{
		{
			name:       "default appliance",
			connection: true,
			envvars: map[string]string{
				"PGHOST":     "pgsql",
				"PGPORT":     "5432",
				"PGUSER":     "sg",
				"PGPASSWORD": "sg",
				"PGDATABASE": "sg",
				"PGSSLMODE":  "disable",
			},
		},
		{
			name:       "malformed dsn",
			connection: false,
			envvars: map[string]string{
				"PGHOST":     "pgsql",
				"PGPORT":     "42069",
				"PGUSER":     "DrDoom",
				"PGPASSWORD": "Doombot",
				"PGDATABASE": "Latveria",
				"PGSSLMODE":  "disable",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.envvars {
				os.Setenv(k, v)
			}
			err := CheckConnection(&observation.TestContext, "postgres")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if test.connection {
				t.Log("connection")
			} else {
				t.Log("no connection")
			}
		})
	}
}
