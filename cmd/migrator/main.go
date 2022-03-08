package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/prometheus/client_golang/prometheus"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const appName = "migrator"

var out = output.NewOutput(os.Stdout, output.OutputOpts{
	ForceColor: true,
	ForceTTY:   true,
})

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = append(args, "up")
	}

	if err := mainErr(context.Background(), args); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr(ctx context.Context, args []string) error {
	runnerFactory := newRunnerFactory()
	rootFlagSet := flag.NewFlagSet(appName, flag.ExitOnError)
	command := &ffcli.Command{
		Name:       appName,
		ShortUsage: fmt.Sprintf("%s <command>", appName),
		ShortHelp:  "Validates and runs schema migrations",
		FlagSet:    rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			cliutil.Up(appName, runnerFactory, out, false),
			cliutil.UpTo(appName, runnerFactory, out, false),
			cliutil.DownTo(appName, runnerFactory, out, false),
			cliutil.Validate(appName, runnerFactory, out),
			cliutil.AddLog(appName, runnerFactory, out),
		},
	}

	if err := command.Parse(args); err != nil {
		return err
	}

	return command.Run(ctx)
}

func newRunnerFactory() func(ctx context.Context, schemaNames []string) (cliutil.Runner, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	operations := store.NewOperations(observationContext)

	return func(ctx context.Context, schemaNames []string) (cliutil.Runner, error) {
		dsns, err := postgresdsn.DSNsBySchema(schemaNames)
		if err != nil {
			return nil, err
		}
		storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
			return connections.NewStoreShim(store.NewWithDB(db, migrationsTable, operations))
		}
		r, err := connections.RunnerFromDSNs(dsns, appName, storeFactory)
		if err != nil {
			return nil, err
		}

		return cliutil.NewShim(r), nil
	}
}
