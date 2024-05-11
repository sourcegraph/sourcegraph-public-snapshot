package config

import (
	"embed"
)

var (
	//go:embed pgsql/postgresql.conf
	fs embed.FS

	pgsqlConfig []byte
)

func init() {
	pgsqlConfig, _ = fs.ReadFile("pgsql/postgresql.conf")
}

func DefaultPGSQLConfig() string {
	return string(pgsqlConfig)
}
