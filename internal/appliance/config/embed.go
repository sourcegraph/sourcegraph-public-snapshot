package config

import (
	"embed"
)

var (
	//go:embed postgres/*
	fs embed.FS

	PgsqlConfig     []byte
	CodeIntelConfig []byte
)

func init() {
	PgsqlConfig, _ = fs.ReadFile("postgres/pgsql.conf")
}

func init() {
	CodeIntelConfig, _ = fs.ReadFile("postgres/codeintel.conf")
}
