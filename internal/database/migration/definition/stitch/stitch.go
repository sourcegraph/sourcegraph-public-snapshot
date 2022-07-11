package stitch

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const squashedMigrationPrefix = "squashed migrations"

func StitchDefinitions(schemaName string, revs []string) (*definition.Definitions, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	migrationsPath := filepath.Join("migrations", schemaName)

	definitionMap := map[int]definition.Definition{}
	for _, rev := range revs {
		revDefinitions, err := definition.ReadDefinitions(newGitFS(rev, root, migrationsPath), migrationsPath)
		if err != nil {
			return nil, errors.Wrap(err, "@"+rev)
		}

		for i, newDefinition := range revDefinitions.All() {
			isSquashedMigration := i <= 1

			// validate assumption that the condition above holds
			if isSquashedMigration && !strings.HasPrefix(newDefinition.Name, squashedMigrationPrefix) {
				return nil, errors.Newf("expected migration %d to have a name prefixed with %q", newDefinition.ID, squashedMigrationPrefix)
			}

			existingDefinition, ok := definitionMap[newDefinition.ID]
			if !ok {
				definitionMap[newDefinition.ID] = newDefinition
				continue
			}

			// Check for edited migrations, but ignore squashed definitions that have reused
			// an old migration definition identifier.
			if !isSquashedMigration && !compareDefinitions(newDefinition, existingDefinition) {
				return nil, errors.Newf("migration %d unexpectedly edited in release %s", newDefinition.ID, rev)
			}
		}
	}

	migrationDefinitions := make([]definition.Definition, 0, len(definitionMap))
	for _, v := range definitionMap {
		migrationDefinitions = append(migrationDefinitions, v)
	}

	return definition.NewDefinitions(migrationDefinitions)
}

func compareDefinitions(x, y definition.Definition) bool {
	return cmp.Diff(x, y, cmp.Comparer(func(x, y *sqlf.Query) bool {
		// Note: migrations do not have args to compare here, so we can compare only
		// the query text safely. If we ever need to add runtime arguments to the
		// migration runner, this assumption _might_ change.
		return x.Query(sqlf.PostgresBindVar) == y.Query(sqlf.PostgresBindVar)
	})) == ""
}
