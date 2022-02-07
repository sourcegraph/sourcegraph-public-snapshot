package schemas

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/migrations"
)

var (
	Frontend     = mustResolveSchema("frontend")
	CodeIntel    = mustResolveSchema("codeintel")
	CodeInsights = mustResolveSchema("codeinsights")

	Schemas = []*Schema{
		Frontend,
		CodeIntel,
		CodeInsights,
	}
)

func mustResolveSchema(name string) *Schema {
	fs, err := fs.Sub(migrations.QueryDefinitions, name)
	if err != nil {
		panic(fmt.Sprintf("malformed migration definitions %q: %s", name, err))
	}

	definitions, err := definition.ReadDefinitions(fs)
	if err != nil {
		panic(fmt.Sprintf("malformed migration definitions %q: %s", name, err))
	}

	return &Schema{
		Name:                name,
		MigrationsTableName: strings.TrimPrefix(fmt.Sprintf("%s_schema_migrations", name), "frontend_"),
		FS:                  fs,
		Definitions:         definitions,
	}
}
