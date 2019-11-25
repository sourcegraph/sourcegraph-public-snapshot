package shared

import "testing"

func TestPostgresExporterProcFile(t *testing.T) {
	env := map[string]string{
		"PGPORT":     "6789",
		"PGHOST":     "scooby",
		"PGUSER":     "mallat",
		"PGPASSWORD": "redbull",
		"PGSSLMODE":  "disable",
	}

	line, err := postgresExporterProcFileWithLooker(func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	})

	if err != nil {
		t.Error(err)
	}

	expected := `postgres_exporter: env DATA_SOURCE_NAME="postgresql://mallat:redbull@scooby:6789/postgres?sslmode=disable" postgres_exporter`
	if line != expected {
		t.Errorf("expected %s, got %s", expected, line)
	}
}
