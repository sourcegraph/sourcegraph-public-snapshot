package testdata

import "embed"

//go:embed well-formed/*.sql query-error/*.sql
var Content embed.FS
