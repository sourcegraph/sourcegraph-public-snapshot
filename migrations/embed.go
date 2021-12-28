package migrations

import "embed"

//go:embed frontend/*.sql codeintel/*.sql codeinsights/*.sql
var QueryDefinitions embed.FS
