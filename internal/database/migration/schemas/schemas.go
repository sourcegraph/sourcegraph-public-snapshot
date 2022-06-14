package schemas

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/migrations"
)

var (
	FrontendDefinition     = mustResolveSchema("frontend")
	CodeIntelDefinition    = mustResolveSchema("codeintel")
	CodeInsightsDefinition = mustResolveSchema("codeinsights")

	Schemas = []*Schema{
		FrontendDefinition,
		CodeIntelDefinition,
		CodeInsightsDefinition,
	}
)

func mustResolveSchema(name string) *Schema {
	fs, err := fs.Sub(migrations.QueryDefinitions, name)
	if err != nil {
		panic(fmt.Sprintf("malformed migration definitions %q: %s", name, err))
	}

	schema, err := ResolveSchema(fs, name)
	if err != nil {
		panic(err.Error())
	}

	return schema
}

func ResolveSchema(fs fs.FS, name string) (*Schema, error) {
	definitions, err := definition.ReadDefinitions(fs, filepath.Join("migrations", name))
	if err != nil {
		return nil, errors.Newf("malformed migration definitions %q: %s", name, err)
	}

	return &Schema{
		Name:                name,
		MigrationsTableName: strings.TrimPrefix(fmt.Sprintf("%s_schema_migrations", name), "frontend_"),
		FS:                  fs,
		Definitions:         definitions,
	}, nil
}
