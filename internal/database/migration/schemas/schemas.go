package schemas

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	fsys, err := fs.Sub(migrations.QueryDefinitions, name)
	if err != nil {
		panic(fmt.Sprintf("malformed migration definitions %q: %s", name, err))
	}

	schema, err := ResolveSchema(fsys, name)
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
		MigrationsTableName: MigrationsTableName(name),
		Definitions:         definitions,
	}, nil
}

func ResolveSchemaAtRev(name, rev string) (*Schema, error) {
	definitions, err := shared.GetFrozenDefinitions(name, rev)
	if err != nil {
		return nil, err
	}

	return &Schema{
		Name:                name,
		MigrationsTableName: MigrationsTableName(name),
		Definitions:         definitions,
	}, nil
}

// MigrationsTableName returns the original name used by golang-migrate. This name has since been used to
// identify each schema uniquely in the same fashion. Maybe someday we'll be able to migrate to just using
// the raw schema name transparently.i
func MigrationsTableName(name string) string {
	return strings.TrimPrefix(fmt.Sprintf("%s_schema_migrations", name), "frontend_")
}

// FilterSchemasByName returns a copy of the given schemas slice containing only schema matching the given
// set of names.
func FilterSchemasByName(schemas []*Schema, targetNames []string) []*Schema {
	filtered := make([]*Schema, 0, len(schemas))
	for _, schema := range schemas {
		for _, targetName := range targetNames {
			if targetName == schema.Name {
				filtered = append(filtered, schema)
				break
			}
		}
	}

	return filtered
}

// getSchemaJSONFilename returns the basename of the JSON-serialized schema in the sg/sg repository.
func GetSchemaJSONFilename(schemaName string) (string, error) {
	switch schemaName {
	case "frontend":
		return "internal/database/schema.json", nil
	case "codeintel":
		fallthrough
	case "codeinsights":
		return fmt.Sprintf("internal/database/schema.%s.json", schemaName), nil
	}

	return "", errors.Newf("unknown schema name %q", schemaName)
}
