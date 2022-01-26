package testdata

import "embed"

//go:embed concurrent-down/*.sql well-formed/*.sql missing-upgrade-query/*.sql missing-downgrade-query/*.sql duplicate-upgrade-query/*.sql duplicate-downgrade-query/*.sql gap-in-sequence/*.sql root-with-parent/*.sql unexpected-parent/*.sql
var Content embed.FS
