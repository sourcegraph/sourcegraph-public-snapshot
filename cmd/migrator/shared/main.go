package shared

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	ossmigrations "github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const appName = "migrator"

var out = output.NewOutput(os.Stdout, output.OutputOpts{})

// NewRunnerWithSchemas returns new migrator runner with given scheme names and
// definitions.
func NewRunnerWithSchemas(observationCtx *observation.Context, logger log.Logger, schemaNames []string, schemas []*schemas.Schema) (cliutil.Runner, error) {
	dsns, err := postgresdsn.DSNsBySchema(schemaNames)
	if err != nil {
		return nil, err
	}

	var dsnsStrings []string
	for schema, dsn := range dsns {
		dsnsStrings = append(dsnsStrings, schema+" => "+dsn)
	}

	out.WriteLine(output.Linef(output.EmojiInfo, output.StyleGrey, "Connection DSNs used: %s", strings.Join(dsnsStrings, ", ")))

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(observationCtx, db, migrationsTable))
	}
	r, err := connections.RunnerFromDSNsWithSchemas(out, logger, dsns, appName, storeFactory, schemas)
	if err != nil {
		return nil, err
	}

	return cliutil.NewShim(r), nil
}

// DefaultSchemaFactories is a list of schema factories to be used in
// non-exceptional cases.
var DefaultSchemaFactories = []cliutil.ExpectedSchemaFactory{
	cliutil.LocalExpectedSchemaFactory,
	cliutil.GitHubExpectedSchemaFactory,
	cliutil.GCSExpectedSchemaFactory,
}

func Start(logger log.Logger, registerEnterpriseMigrators registerMigratorsUsingConfAndStoreFactoryFunc) error {
	observationCtx := observation.NewContext(logger)

	outputFactory := func() *output.Output { return out }

	newRunnerWithSchemas := func(schemaNames []string, schemas []*schemas.Schema) (cliutil.Runner, error) {
		return NewRunnerWithSchemas(observationCtx, logger, schemaNames, schemas)
	}
	newRunner := func(schemaNames []string) (cliutil.Runner, error) {
		return newRunnerWithSchemas(schemaNames, schemas.Schemas)
	}

	registerMigrators := composeRegisterMigratorsFuncs(
		ossmigrations.RegisterOSSMigratorsUsingConfAndStoreFactory,
		registerEnterpriseMigrators,
	)

	command := &cli.App{
		Name:   appName,
		Usage:  "Validates and runs schema migrations",
		Action: cli.ShowSubcommandHelp,
		Commands: []*cli.Command{
			cliutil.Up(appName, newRunner, outputFactory, false),
			cliutil.UpTo(appName, newRunner, outputFactory, false),
			cliutil.DownTo(appName, newRunner, outputFactory, false),
			cliutil.Validate(appName, newRunner, outputFactory),
			cliutil.Describe(appName, newRunner, outputFactory),
			cliutil.Drift(appName, newRunner, outputFactory, false, DefaultSchemaFactories...),
			cliutil.AddLog(appName, newRunner, outputFactory),
			cliutil.Upgrade(appName, newRunnerWithSchemas, outputFactory, registerMigrators, DefaultSchemaFactories...),
			cliutil.Downgrade(appName, newRunnerWithSchemas, outputFactory, registerMigrators, DefaultSchemaFactories...),
			cliutil.RunOutOfBandMigrations(appName, newRunner, outputFactory, registerMigrators),
		},
	}

	out.WriteLine(output.Linef(output.EmojiAsterisk, output.StyleReset, "Sourcegraph migrator %s", version.Version()))

	args := os.Args
	if len(args) == 1 {
		args = append(args, "up")
	}

	return command.RunContext(context.Background(), args)
}
