package operator

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// Validate cli code
// - setup runner
//   - takes a runncerFactory which is called to with schemaNames input to return a runner
//     -
// - usees runner validate
// - prints line to container output
// - checks for out of bound migrations

// Create a readonly connection to the databases and check that they're schemas are in the expected state.
// Validate that the expected definitions have been defined
func Validate(version *semver.Version) error {
	ctx := context.Background()
	logger := log.Scoped("appliance")
	observationCtx := observation.NewContext(logger)
	fmt.Println(observationCtx)
	fmt.Println(version)
	fmt.Println(strings.Join(schemaNames, ""))

	// FetchExpectedSchemas

	// get dsn handles on database, handle for local development
	os.Setenv("CODEINTEL_PG_ALLOW_SINGLE_DB", "true")
	dsns, err := postgresdsn.DSNsBySchema(schemaNames)
	if err != nil {
		return err
	}
	fmt.Println(dsns)

	out := output.NewOutput(os.Stdout, output.OutputOpts{})
	newRunnerWithSchemas := func(schemaNames []string, schemas []*schemas.Schema) (*runner.Runner, error) {
		return migration.NewRunnerWithSchemas(observationCtx, out, "migrator", schemaNames, schemas)
	}
	runner, err := newRunnerWithSchemas(schemaNames, schemas.Schemas)
	if err = runner.Validate(ctx, schemaNames...); err != nil {
		return err
	}

	return nil
}

var schemaNames = []string{
	"frontend",
	"codeintel",
	"codeinsights",
}

// func CopyCat() error {
// 	r, err := setupRunner(factory, schemaNames...)
// 	if err != nil {
// 		return err
// 	}

// 	if err := r.Validate(ctx, schemaNames...); err != nil {
// 		return err
// 	}

// 	out.WriteLine(output.Emoji(output.EmojiSuccess, "schema okay!"))

// 	if !skipOutOfBandMigrationsFlag.Get(cmd) {
// 		db, err := store.ExtractDatabase(ctx, r)
// 		if err != nil {
// 			return err
// 		}

// 		if err := oobmigration.ValidateOutOfBandMigrationRunner(ctx, db, outOfBandMigrationRunner(db)); err != nil {
// 			return err
// 		}

// 		out.WriteLine(output.Emoji(output.EmojiSuccess, "oobmigrations okay!"))
// 	}

// 	return nil
// }
