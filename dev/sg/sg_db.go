package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
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
		Name:     "db",
		Usage:    "Interact with local Sourcegraph databases for development",
		Category: CategoryDev,
		Subcommands: []*cli.Command{
			{
				Name:        "reset-pg",
				Usage:       "Drops, recreates and migrates the specified Sourcegraph database",
				Description: `Run 'sg db reset-pg' to drop and recreate Sourcegraph databases. If -db is not set, then the "frontend" database is used (what's set as PGDATABASE in env or the sg.config.yaml). If -db is set to "all" then all databases are reset and recreated.`,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "db",
						Value:       db.DefaultDatabase.Name,
						Usage:       "The target database instance.",
						Destination: &dbDatabaseNameFlag,
					},
				},
				Action: execAdapter(dbResetPGExec),
			},
			{
				Name:        "reset-redis",
				Usage:       "Drops, recreates and migrates the specified Sourcegraph Redis database",
				Description: `Run 'sg db reset-redis' to drop and recreate Sourcegraph redis databases.`,
				Action:      execAdapter(dbResetRedisExec),
			},
			{
				Name:        "add-user",
				Usage:       "Create an admin sourcegraph user",
				Description: `Run 'sg db add-user -name bob' to create an admin user whose email is bob@sourcegraph.com. The password will be printed if the operation succeeds`,
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

	// Read the configuration.
	conf, _ := sgconf.Get(configFile, configOverwriteFile)
	if conf == nil {
		return errors.New("failed to read sg.config.yaml. This command needs to be run in the `sourcegraph` repository")
	}

	// Connect to the database.
	conn, err := connections.EnsureNewFrontendDB(postgresdsn.New("", "", conf.GetEnv), "frontend", &observation.TestContext)
	if err != nil {
		return err
	}
	db := database.NewDB(conn)

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

func dbResetRedisExec(ctx context.Context, args []string) error {
	// Read the configuration.
	config, _ := sgconf.Get(configFile, configOverwriteFile)
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

func dbResetPGExec(ctx context.Context, args []string) error {
	// Read the configuration.
	config, _ := sgconf.Get(configFile, configOverwriteFile)
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

	for name, dsn := range dsnMap {
		var (
			db  *pgx.Conn
			err error
		)

		db, err = pgx.Connect(ctx, dsn)
		if err != nil {
			return errors.Wrap(err, "failed to connect to Postgres database")
		}

		std.Out.WriteNoticef("This will reset database %s%s%s. Are you okay with this?", output.StyleOrange, name, output.StyleReset)
		ok := getBool()
		if !ok {
			return nil
		}

		_, err = db.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			std.Out.WriteFailuref("Failed to drop schema 'public': %s", err)
			return err
		}

		if err := db.Close(ctx); err != nil {
			return err
		}
	}

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(db, migrationsTable, store.NewOperations(&observation.TestContext)))
	}
	r, err := connections.RunnerFromDSNs(dsnMap, "sg", storeFactory)
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

	return r.Run(ctx, runner.Options{
		Operations: operations,
	})
}
