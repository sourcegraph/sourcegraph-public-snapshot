package upgrades

import (
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
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

func TestCheckConnection_Ping(t *testing.T) {
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
			name:       "malformed dsn",
			schema:     "doombot",
			connection: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := dbtest.NewDB(t)
			defer db.Close()

			var currentUser string
			err := db.QueryRow("SELECT current_user").Scan(&currentUser)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				t.FailNow()
			}

			url, err := dbtest.GetDSN()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				t.FailNow()
			}
			password, _ := url.User.Password()

			var dbName string
			err = db.QueryRow("SELECT current_database()").Scan(&dbName)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				t.FailNow()
			}

			t.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")
			t.Setenv("PGUSER", currentUser)
			t.Setenv("PGPASSWORD", password)
			t.Setenv("PGDATABASE", dbName)
			t.Setenv("PGSSLMODE", "disable")
			t.Setenv("PGTZ", "UTC")

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

// TODO: This should check that the connections with no envvars set remain as default,
// theres some odd handling of defaults setting current user in the postgresdsns package
// this may need further investigation.
func TestGetApplianceDSNs(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		envvars     map[string]string
		expectedDSN string
	}{
		{
			name:   "default frontend",
			schema: "frontend",
			envvars: map[string]string{
				"PGHOST":     "pgsql",
				"PGPORT":     "5432",
				"PGUSER":     "sg",
				"PGPASSWORD": "sg",
				"PGDATABASE": "sg",
				"PGSSLMODE":  "disable",
				"PGTZ":       "UTC",
			},
			expectedDSN: "postgres://sg:sg@pgsql:5432/sg?sslmode=disable&timezone=UTC",
		},
		{
			name:   "default codeintel",
			schema: "codeintel",
			envvars: map[string]string{
				"CODEINTEL_PGHOST":     "codeintel-db",
				"CODEINTEL_PGPORT":     "5432",
				"CODEINTEL_PGUSER":     "sg",
				"CODEINTEL_PGPASSWORD": "sg",
				"CODEINTEL_PGDATABASE": "sg",
				"CODEINTEL_PGSSLMODE":  "disable",
				"CODEINTEL_PGTZ":       "UTC",
			},
			expectedDSN: "postgres://sg:sg@codeintel-db:5432/sg?sslmode=disable&timezone=UTC",
		},
		{
			name:   "default codeinsights",
			schema: "codeinsights",
			envvars: map[string]string{
				"CODEINSIGHTS_PGHOST":     "codeinsights-db",
				"CODEINSIGHTS_PGPORT":     "5432",
				"CODEINSIGHTS_PGUSER":     "postgres",
				"CODEINSIGHTS_PGPASSWORD": "password",
				"CODEINSIGHTS_PGDATABASE": "postgres",
				"CODEINSIGHTS_PGSSLMODE":  "disable",
				"CODEINSIGHTS_PGTZ":       "UTC",
			},
			expectedDSN: "postgres://postgres:password@codeinsights-db:5432/postgres?sslmode=disable&timezone=UTC",
		},
		{
			name:   "unusual dsn",
			schema: "frontend",
			envvars: map[string]string{
				"PGHOST":     "latveria",
				"PGPORT":     "6969",
				"PGUSER":     "doombot",
				"PGPASSWORD": "allhaildoom",
				"PGDATABASE": "postgres",
				"PGSSLMODE":  "disable",
				"PGTZ":       "UTC",
			},
			expectedDSN: "postgres://doombot:allhaildoom@latveria:6969/postgres?sslmode=disable&timezone=UTC",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.envvars {
				os.Setenv(k, v)
			}

			dsns, err := getApplianceDSNs()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				t.FailNow()
			}
			t.Log("dsn: ", dsns)

			if dsns[test.schema] != test.expectedDSN {
				t.Errorf("expected dsn %s, got %s", test.expectedDSN, dsns[test.schema])
			}
		})
	}
}
