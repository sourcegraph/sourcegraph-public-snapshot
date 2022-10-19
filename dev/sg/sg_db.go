package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/sg/cliutil"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	dbDatabaseNameFlag string

	dbCommand = &cli.Command{
		Name:  "db",
		Usage: "Interact with local Sourcegraph databases for development",
		UsageText: `
# Delete test databases
sg db delete-test-dbs

# Reset the Sourcegraph 'frontend' database
sg db reset-pg

# Reset the 'frontend' and 'codeintel' databases
sg db reset-pg -db=frontend,codeintel

# Reset all databases ('frontend', 'codeintel', 'codeinsights')
sg db reset-pg -db=all

# Reset the redis database
sg db reset-redis

# Create a site-admin user whose email and password are foo@sourcegraph.com and sourcegraph.
sg db add-user -name=foo
`,
		Category: CategoryDev,
		Subcommands: []*cli.Command{
			{
				Name:   "delete-test-dbs",
				Usage:  "Drops all databases that have the prefix `sourcegraph-test-`",
				Action: deleteTestDBsExec,
			},
			{
				Name:        "reset-pg",
				Usage:       "Drops, recreates and migrates the specified Sourcegraph database",
				Description: `If -db is not set, then the "frontend" database is used (what's set as PGDATABASE in env or the sg.config.yaml). If -db is set to "all" then all databases are reset and recreated.`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "db",
						Value:       db.DefaultDatabase.Name,
						Usage:       "The target database instance.",
						Destination: &dbDatabaseNameFlag,
					},
				},
				Action: dbResetPGExec,
			},
			{
				Name:      "reset-redis",
				Usage:     "Drops, recreates and migrates the specified Sourcegraph Redis database",
				UsageText: "sg db reset-redis",
				Action:    dbResetRedisExec,
			},
			{
				Name:        "add-user",
				Usage:       "Create an admin sourcegraph user",
				Description: `Run 'sg db add-user -username bob' to create an admin user whose email is bob@sourcegraph.com. The password will be printed if the operation succeeds`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "username",
						Value: "sourcegraph",
						Usage: "Username for user",
					},
					&cli.StringFlag{
						Name:  "password",
						Value: "sourcegraphsourcegraph",
						Usage: "Password for user",
					},
				},
				Action: dbAddUserAction,
			},
		},
	}
)

func dbAddUserAction(cmd *cli.Context) error {
	ctx := cmd.Context
	logger := log.Scoped("dbAddUserAction", "")

	// Read the configuration.
	conf, _ := getConfig()
	if conf == nil {
		return errors.New("failed to read sg.config.yaml. This command needs to be run in the `sourcegraph` repository")
	}

	// Connect to the database.
	conn, err := connections.EnsureNewFrontendDB(postgresdsn.New("", "", conf.GetEnv), "frontend", &observation.TestContext)
	if err != nil {
		return err
	}
	db := database.NewDB(logger, conn)

	username := cmd.String("username")
	password := cmd.String("password")

	// Create the user, generating an email based on the username.
	email := fmt.Sprintf("%s@sourcegraph.com", username)
	user, err := db.Users().Create(ctx, database.NewUser{
		Username:        username,
		Email:           email,
		EmailIsVerified: true,
		Password:        password,
	})
	if err != nil {
		return err
	}

	// Make the user site admin.
	err = db.Users().SetIsSiteAdmin(ctx, user.ID, true)
	if err != nil {
		return err
	}

	// Report back the new user information.
	std.Out.WriteSuccessf(
		// the space after the last %s is so the user can select the password easily in the shell to copy it.
		"User '%s%s%s' (%s%s%s) has been created and its password is '%s%s%s'.",
		output.StyleOrange,
		username,
		output.StyleReset,
		output.StyleOrange,
		email,
		output.StyleReset,
		output.StyleOrange,
		password,
		output.StyleReset,
	)

	return nil
}

func dbResetRedisExec(ctx *cli.Context) error {
	// Read the configuration.
	config, _ := getConfig()
	if config == nil {
		return errors.New("failed to read sg.config.yaml. This command needs to be run in the `sourcegraph` repository")
	}

	// Connect to the redis database.
	endpoint := config.GetEnv("REDIS_ENDPOINT")
	conn, err := redis.Dial("tcp", endpoint, redis.DialConnectTimeout(5*time.Second))
	if err != nil {
		return errors.Wrapf(err, "failed to connect to Redis at %s", endpoint)
	}

	// Drop everything in redis
	_, err = conn.Do("flushall")
	if err != nil {
		return errors.Wrap(err, "failed to run command on redis")
	}

	return nil
}

func deleteTestDBsExec(ctx *cli.Context) error {
	config, err := dbtest.GetDSN()
	if err != nil {
		return err
	}
	dsn := config.String()

	db, err := dbconn.ConnectInternal(log.Scoped("sg", ""), dsn, "", "")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = errors.Append(err, closeErr)
			}
		}
	}()

	names, err := basestore.ScanStrings(db.QueryContext(ctx.Context, `SELECT datname FROM pg_database WHERE datname LIKE 'sourcegraph-test-%'`))
	if err != nil {
		return err
	}

	for _, name := range names {
		_, err := db.ExecContext(ctx.Context, fmt.Sprintf(`DROP DATABASE %q`, name))
		if err != nil {
			return err
		}

		std.Out.WriteLine(output.Linef(output.EmojiOk, output.StyleReset, fmt.Sprintf("Deleted %s", name)))
	}

	std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, fmt.Sprintf("%d databases deleted.", len(names))))
	return nil
}

func dbResetPGExec(ctx *cli.Context) error {
	// Read the configuration.
	config, _ := getConfig()
	if config == nil {
		return errors.New("failed to read sg.config.yaml. This command needs to be run in the `sourcegraph` repository")
	}

	var (
		dsnMap      = map[string]string{}
		schemaNames []string
	)

	if dbDatabaseNameFlag == "all" {
		schemaNames = schemas.SchemaNames
	} else {
		schemaNames = strings.Split(dbDatabaseNameFlag, ",")
	}

	for _, name := range schemaNames {
		if name == "frontend" {
			dsnMap[name] = postgresdsn.New("", "", config.GetEnv)
		} else {
			dsnMap[name] = postgresdsn.New(strings.ToUpper(name), "", config.GetEnv)
		}
	}

	std.Out.WriteNoticef("This will reset database(s) %s%s%s. Are you okay with this?",
		output.StyleOrange, strings.Join(schemaNames, ", "), output.StyleReset)
	if ok := getBool(); !ok {
		return cliutil.NewEmptyExitErr(1)
	}

	for _, dsn := range dsnMap {
		var (
			db  *pgx.Conn
			err error
		)

		db, err = pgx.Connect(ctx.Context, dsn)
		if err != nil {
			return errors.Wrap(err, "failed to connect to Postgres database")
		}

		_, err = db.Exec(ctx.Context, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			std.Out.WriteFailuref("Failed to drop schema 'public': %s", err)
			return err
		}

		if err := db.Close(ctx.Context); err != nil {
			return err
		}
	}

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(db, migrationsTable, store.NewOperations(&observation.TestContext)))
	}
	r, err := connections.RunnerFromDSNs(log.Scoped("migrations.runner", ""), dsnMap, "sg", storeFactory)
	if err != nil {
		return err
	}

	operations := make([]runner.MigrationOperation, 0, len(schemaNames))
	for _, schemaName := range schemaNames {
		operations = append(operations, runner.MigrationOperation{
			SchemaName: schemaName,
			Type:       runner.MigrationOperationTypeUpgrade,
		})
	}

	if err := r.Run(ctx.Context, runner.Options{
		Operations: operations,
	}); err != nil {
		return err
	}

	std.Out.WriteSuccessf("Database(s) reset!")
	return nil
}
