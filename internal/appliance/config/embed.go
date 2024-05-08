package config

import (
	"embed"
)

var (
	//go:embed pgsql/postgresql.conf
	fs embed.FS

	postgresqlConfig []byte
)

func init() {
	postgresqlConfig, _ = fs.ReadFile("pgsql/postgresql.conf")
}

func PostgresqlConfig() string {
	return string(postgresqlConfig)
}
