package shared

import (
	"context"
	"database/sql"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	ossmigrations "github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const appName = "migrator"

var out = output.NewOutput(os.Stdout, output.OutputOpts{
	ForceColor: true,
	ForceTTY:   true,
})

func Start(logger log.Logger, registerEnterpriseMigrators registerMigratorsUsingConfAndStoreFactoryFunc) error {
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}
	operations := store.NewOperations(observationContext)

	outputFactory := func() *output.Output { return out }

	newRunnerWithSchemas := func(ctx context.Context, schemaNames []string, schemas []*schemas.Schema) (cliutil.Runner, error) {
		dsns, err := postgresdsn.DSNsBySchema(schemaNames)
		if err != nil {
			return nil, err
		}
		storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
			return connections.NewStoreShim(store.NewWithDB(db, migrationsTable, operations))
		}
		r, err := connections.RunnerFromDSNsWithSchemas(logger, dsns, appName, storeFactory, schemas)
		if err != nil {
			return nil, err
		}

		return cliutil.NewShim(r), nil
	}
	newRunner := func(ctx context.Context, schemaNames []string) (cliutil.Runner, error) {
		return newRunnerWithSchemas(ctx, schemaNames, schemas.Schemas)
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
			cliutil.Drift(appName, newRunner, outputFactory, cliutil.GCSExpectedSchemaFactory, cliutil.GitHubExpectedSchemaFactory),
			cliutil.AddLog(appName, newRunner, outputFactory),
			cliutil.Upgrade(appName, newRunnerWithSchemas, outputFactory, registerMigrators),
			cliutil.Downgrade(appName, newRunnerWithSchemas, outputFactory, registerMigrators),
			cliutil.RunOutOfBandMigrations(appName, newRunner, outputFactory, registerMigrators),
		},
	}

	out.WriteLine(output.Linef(output.EmojiAsterisk, output.StyleReset, "Sourcegraph migrator v%s", version.Version()))

	args := os.Args
	if len(args) == 1 {
		args = append(args, "up")
	}

	return command.RunContext(context.Background(), args)
}
