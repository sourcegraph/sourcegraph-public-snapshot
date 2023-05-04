package cliutil

import (
	"context"
	"database/sql"
	"flag"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type actionFunction func(ctx context.Context, cmd *cli.Context, out *output.Output) error

// makeAction creates a new migration action function. It is expected that these
// commands accept zero arguments and define their own flags.
func makeAction(outFactory OutputFactory, f actionFunction) func(cmd *cli.Context) error {
	return func(cmd *cli.Context) error {
		if cmd.NArg() != 0 {
			return flagHelp(outFactory(), "too many arguments")
		}

		return f(cmd.Context, cmd, outFactory())
	}
}

// flagHelp returns an error that prints the specified error message with usage text.
func flagHelp(out *output.Output, message string, args ...any) error {
	out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: "+message, args...))
	return flag.ErrHelp
}

// setupRunner initializes and returns the runner associated witht the given schema.
func setupRunner(factory RunnerFactory, schemaNames ...string) (Runner, error) {
	r, err := factory(schemaNames)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// setupStore initializes and returns the store associated witht the given schema.
func setupStore(ctx context.Context, factory RunnerFactory, schemaName string) (Store, error) {
	r, err := setupRunner(factory, schemaName)
	if err != nil {
		return nil, err
	}

	store, err := r.Store(ctx, schemaName)
	if err != nil {
		return nil, err
	}

	return store, nil
}

// sanitizeSchemaNames sanitizies the given string slice from the user.
func sanitizeSchemaNames(schemaNames []string, out *output.Output) []string {
	if len(schemaNames) == 1 && schemaNames[0] == "" {
		schemaNames = nil
	}

	if len(schemaNames) == 1 && schemaNames[0] == "all" {
		return schemas.SchemaNames
	}

	for i, name := range schemaNames {
		schemaNames[i] = TranslateSchemaNames(name, out)
	}

	return schemaNames
}

var dbNameToSchema = map[string]string{
	"pgsql":           "frontend",
	"codeintel-db":    "codeintel",
	"codeinsights-db": "codeinsights",
}

// TranslateSchemaNames translates a string with potentially the value of the service/container name
// of the db schema the user wants to operate on into the schema name.
func TranslateSchemaNames(name string, out *output.Output) string {
	// users might input the name of the service e.g. pgsql instead of frontend, so we
	// translate to what it actually should be
	if translated, ok := dbNameToSchema[name]; ok {
		out.WriteLine(output.Linef(output.EmojiInfo, output.StyleGrey, "Translating container/service name %q to schema name %q", name, translated))
		name = translated
	}

	return name
}

// parseTargets parses the given strings as integers.
func parseTargets(targets []string) ([]int, error) {
	if len(targets) == 1 && targets[0] == "" {
		targets = nil
	}

	versions := make([]int, 0, len(targets))
	for _, target := range targets {
		version, err := strconv.Atoi(target)
		if err != nil {
			return nil, err
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// getPivilegedModeFromFlags transforms the given flags into an equivalent PrivilegedMode value. A user error is
// returned if the supplied flags form an invalid state.
func getPivilegedModeFromFlags(cmd *cli.Context, out *output.Output, unprivilegedOnlyFlag, noopPrivilegedFlag *cli.BoolFlag) (runner.PrivilegedMode, error) {
	unprivilegedOnly := unprivilegedOnlyFlag.Get(cmd)
	noopPrivileged := noopPrivilegedFlag.Get(cmd)
	if unprivilegedOnly && noopPrivileged {
		return runner.InvalidPrivilegedMode, flagHelp(out, "-unprivileged-only and -noop-privileged are mutually exclusive")
	}

	if unprivilegedOnly {
		return runner.RefusePrivilegedMigrations, nil
	}
	if noopPrivileged {
		return runner.NoopPrivilegedMigrations, nil
	}

	return runner.ApplyPrivilegedMigrations, nil
}

func extractDatabase(ctx context.Context, r Runner) (database.DB, error) {
	db, err := extractDB(ctx, r, "frontend")
	if err != nil {
		return nil, err
	}

	return database.NewDB(log.Scoped("migrator", ""), db), nil
}

func extractDB(ctx context.Context, r Runner, schemaName string) (*sql.DB, error) {
	store, err := r.Store(ctx, schemaName)
	if err != nil {
		return nil, err
	}

	// NOTE: The migration runner package cannot import basestore without
	// creating a cyclic import in db connection packages. Hence, we cannot
	// embed basestore.ShareableStore here and must "backdoor" extract the
	// database connection.
	shareableStore, ok := basestore.Raw(store)
	if !ok {
		return nil, errors.New("store does not support direct database handle access")
	}

	return shareableStore, nil
}

var migratorObservationCtx = &observation.TestContext

func outOfBandMigrationRunner(db database.DB) *oobmigration.Runner {
	return oobmigration.NewRunnerWithDB(migratorObservationCtx, db, time.Second)
}

func outOfBandMigrationRunnerWithStore(store *oobmigration.Store) *oobmigration.Runner {
	return oobmigration.NewRunner(migratorObservationCtx, store, time.Second)
}
