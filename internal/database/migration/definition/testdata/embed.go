package testdata

import "embed"

//go:embed well-formed/*.sql missing-upgrade-query/*.sql missing-downgrade-query/*.sql duplicate-upgrade-query/*.sql duplicate-downgrade-query/*.sql gap-in-sequence/*.sql
var Content embed.FS
