package main

import (
	"context"
	"database/sql"
	"flag"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"github.com/peterbourgon/ff/v3/ffcli"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var (
	resetFlagSet = flag.NewFlagSet("sg reset", flag.ExitOnError)
	resetCommand = &ffcli.Command{
		Name:       "reset",
		ShortUsage: "sg reset",
		ShortHelp:  "Resets stuff.",
		LongHelp:   `Resets stuff`,
		FlagSet:    resetFlagSet,
		Exec:       resetExec,
	}
)

func resetExec(ctx context.Context, args []string) error {
	getEnv := func(key string) string {
		// First look into process env, emulating the logic in makeEnv used
		// in internal/run/run.go
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
		// Otherwise check in globalConf.Env
		return globalConf.Env[key]
	}

	dsnMap := map[string]string{}
	for _, name := range schemas.SchemaNames {
		if name == "frontend" {
			dsnMap[name] = postgresdsn.New("", "", getEnv)
		} //  else {
		// 	dsnMap[name] = postgresdsn.New(strings.ToUpper(name), "", getEnv)
		// }

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

		writeFingerPointingLine("This will reset database %s. Are you okay with this?", name)
		ok := getBool()
		if !ok {
			return nil
		}

		_, err = db.Exec(ctx, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			writeFailureLine("Failed to drop schema 'public': %s", err)
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
		SchemaNames:   []string{"frontend"},
	}

	return connections.RunnerFromDSNs(dsnMap, "sg", storeFactory).Run(ctx, options)
}
