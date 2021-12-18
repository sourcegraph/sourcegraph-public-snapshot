package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/garyburd/redigo/redis"
	"github.com/jackc/pgx/v4"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	resetFlagSet             = flag.NewFlagSet("sg reset", flag.ExitOnError)
	resetPGFlagSet           = flag.NewFlagSet("sg reset pg", flag.ExitOnError)
	resetDatabaseNameFlag    = resetPGFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance.")
	resetRedisFlagSet        = flag.NewFlagSet("sg reset redis", flag.ExitOnError)
	resetAddUserFlagSet      = flag.NewFlagSet("sg reset add-user", flag.ExitOnError)
	resetAddUserNameFlag     = resetAddUserFlagSet.String("name", "sourcegraph", "User name")
	resetAddUserPasswordFlag = resetAddUserFlagSet.String("password", "sourcegraph", "User password")

	resetCommand = &ffcli.Command{
		Name:       "reset",
		ShortUsage: "",
		LongHelp:   "",
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			{
				Name:       "pg",
				ShortUsage: fmt.Sprintf("sg reset [-db=%s]", db.DefaultDatabase.Name),
				ShortHelp:  "Drops, recreates and migrates the specified Sourcegraph database.",
				LongHelp:   `Run 'sg reset' to drop and recreate Sourcegraph databases. If -db is not set, then the "frontend" database is used (what's set as PGDATABASE in env or the sg.config.yaml). If -db is set to "all" then all databases are reset and recreated.`,
				FlagSet:    resetPGFlagSet,
				Exec:       resetPGExec,
			},
			{
				Name:       "redis",
				ShortUsage: fmt.Sprintf("sg reset redis [-db=%s]", db.DefaultDatabase.Name), // TODO edit flag
				ShortHelp:  "Drops, recreates and migrates the specified redis Sourcegraph database.",
				LongHelp:   `Run 'sg reset redis' to drop and recreate Sourcegraph redis databases. TODO`,
				FlagSet:    resetRedisFlagSet,
				Exec:       resetRedisExec,
			},
			{
				Name:       "add-user",
				ShortUsage: fmt.Sprintf("sg reset add-user [-name=%s -password=%s]", "sourcegraph", "sourcegraphsourcegraph"), // TODO edit flag
				ShortHelp:  "Create a sourcegraph user",
				LongHelp:   `TODO`,
				FlagSet:    resetAddUserFlagSet,
				Exec:       resetAddUserExec,
			},
		},
	}
)

func resetAddUserExec(ctx context.Context, args []string) error {
	conn, err := connections.NewFrontendDB("", "frontend", true, &observation.TestContext)
	if err != nil {
		return err
	}
	_ = database.NewDB(conn)
	time.Sleep(5 * time.Second)
	fmt.Println("fofof|")
	// _, err = db.Users().Create(ctx, database.NewUser{
	// 	Username:        *resetAddUserNameFlag,
	// 	Email:           fmt.Sprintf("%s@sourcegraph.com", *resetDatabaseNameFlag),
	// 	EmailIsVerified: true,
	// 	Password:        *resetAddUserPasswordFlag,
	// })
	// time.Sleep(5 * time.Second)
	// if err != nil {
	// 	return err
	// }

	// err = db.Users().SetIsSiteAdmin(ctx, user.ID, true)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(user.ID)
	return nil
}

func resetRedisExec(ctx context.Context, args []string) error {
	ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		return errors.New("failed to read sg.config.yaml. This step of `sg setup` needs to be run in the `sourcegraph` repository")
	}

	endpoint := globalConf.Env["REDIS_ENDPOINT"]

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

func resetPGExec(ctx context.Context, args []string) error {
	ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
	if !ok {
		return errors.New("failed to read sg.config.yaml. This step of `sg setup` needs to be run in the `sourcegraph` repository")
	}

	getEnv := func(key string) string {
		// First look into process env, emulating the logic in makeEnv used
		// in internal/run/run.go
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
		// Otherwise check in globalConf.Env and *expand* the key, because a value might refer to another env var.
		return os.Expand(globalConf.Env[key], func(lookup string) string {
			if lookup == key {
				return os.Getenv(lookup)
			}

			if e, ok := globalConf.Env[lookup]; ok {
				return e
			}
			return os.Getenv(lookup)
		})
	}

	var (
		dsnMap      = map[string]string{}
		schemaNames []string
	)

	if *resetDatabaseNameFlag == "all" {
		schemaNames = schemas.SchemaNames
	} else {
		schemaNames = strings.Split(*resetDatabaseNameFlag, ",")
	}

	for _, name := range schemaNames {
		if name == "frontend" {
			dsnMap[name] = postgresdsn.New("", "", getEnv)
		} else {
			dsnMap[name] = postgresdsn.New(strings.ToUpper(name), "", getEnv)
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

		writeFingerPointingLinef("This will reset database %s%s%s. Are you okay with this?", output.StyleOrange, name, output.StyleReset)
		ok := getBool()
		if !ok {
			return nil
		}

		_, err = db.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			writeFailureLinef("Failed to drop schema 'public': %s", err)
			return err
		}

		if err := db.Close(ctx); err != nil {
			return err
		}
	}

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return store.NewWithDB(db, migrationsTable, store.NewOperations(&observation.TestContext))
	}

	options := runner.Options{
		Up:            true,
		NumMigrations: 0,
		SchemaNames:   schemaNames,
	}

	return connections.RunnerFromDSNs(dsnMap, "sg", storeFactory).Run(ctx, options)
}
