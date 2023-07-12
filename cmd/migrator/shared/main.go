package shared

import (
	"context"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations/register"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const appName = "migrator"

var out = output.NewOutput(os.Stdout, output.OutputOpts{})

func Start(logger log.Logger, registerEnterpriseMigrators store.RegisterMigratorsUsingConfAndStoreFactoryFunc) error {
	observationCtx := observation.NewContext(logger)

	outputFactory := func() *output.Output { return out }

	newRunnerWithSchemas := func(schemaNames []string, schemas []*schemas.Schema) (*runner.Runner, error) {
		return migration.NewRunnerWithSchemas(observationCtx, out, "migrator", schemaNames, schemas)
	}
	newRunner := func(schemaNames []string) (*runner.Runner, error) {
		return newRunnerWithSchemas(schemaNames, schemas.Schemas)
	}

	registerMigrators := store.ComposeRegisterMigratorsFuncs(
		register.RegisterOSSMigratorsUsingConfAndStoreFactory,
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
			cliutil.Drift(appName, newRunner, outputFactory, false, schemas.DefaultSchemaFactories...),
			cliutil.AddLog(appName, newRunner, outputFactory),
			cliutil.Upgrade(appName, newRunnerWithSchemas, outputFactory, registerMigrators, schemas.DefaultSchemaFactories...),
			cliutil.Downgrade(appName, newRunnerWithSchemas, outputFactory, registerMigrators, schemas.DefaultSchemaFactories...),
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
