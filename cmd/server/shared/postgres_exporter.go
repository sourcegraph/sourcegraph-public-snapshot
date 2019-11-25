package shared

import (
	"fmt"
	"os"
)

type envLookerFn func(string) (string, bool)

func getEnv(looker envLookerFn, key string, defaultVal string) string {
	val, ok := looker(key)
	if ok {
		return val
	}
	return defaultVal
}

func postgresExporterProcFile() (string, error) {
	return postgresExporterProcFileWithLooker(os.LookupEnv)
}

func postgresExporterProcFileWithLooker(looker envLookerFn) (string, error) {
	pgUser := getEnv(looker, "PGUSER", "postgres")
	pgPassword := getEnv(looker, "PGPASSWORD", "sourcegraph")
	pgHost := getEnv(looker, "PGHOST", "127.0.0.1")
	pgPort := getEnv(looker, "PGPORT", "5432")
	pgSSLMode := getEnv(looker, "PGSSLMODE", "disable")

	return fmt.Sprintf(`postgres_exporter: env DATA_SOURCE_NAME="postgresql://%s:%s@%s:%s/postgres?sslmode=%s" postgres_exporter`,
		pgUser, pgPassword, pgHost, pgPort, pgSSLMode), nil
}
