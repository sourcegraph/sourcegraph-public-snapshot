package dbconn

import "testing"

func TestBuildConnectionString(t *testing.T) {
	tests := []struct {
		name                   string
		dataSource             string
		wantedConnectionString string
	}{
		{
			name:                   "empty dataSource",
			dataSource:             "",
			wantedConnectionString: " fallback_application_name=sourcegraph",
		}, {
			name:                   "connection string",
			dataSource:             "dbname=sourcegraph host=localhost sslmode=verify-full user=sourcegraph",
			wantedConnectionString: "dbname=sourcegraph host=localhost sslmode=verify-full user=sourcegraph fallback_application_name=sourcegraph",
		}, {
			name:                   "postgres URL",
			dataSource:             "postgres://sourcegraph@localhost/sourcegraph?sslmode=verify-full",
			wantedConnectionString: "dbname=sourcegraph host=localhost sslmode=verify-full user=sourcegraph fallback_application_name=sourcegraph",
		}, {
			name:                   "invalid URL",
			dataSource:             "invalid string",
			wantedConnectionString: "invalid string fallback_application_name=sourcegraph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildConnectionString(tt.dataSource); got != tt.wantedConnectionString {
				t.Errorf("buildConnectionString() = %v, want %v", got, tt.wantedConnectionString)
			}
		})
	}
}
