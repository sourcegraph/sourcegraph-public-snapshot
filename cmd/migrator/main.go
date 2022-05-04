package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	sglog "github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const appName = "migrator"

var out = output.NewOutput(os.Stdout, output.OutputOpts{
	ForceColor: true,
	ForceTTY:   true,
})

func main() {
	args := os.Args[:]
	if len(args) == 1 {
		args = append(args, "up")
	}

	if err := mainErr(context.Background(), args); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr(ctx context.Context, args []string) error {
	syncLogs := sglog.Init(sglog.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer syncLogs()

	runnerFactory := newRunnerFactory()
	command := &cli.App{
		Name:   appName,
		Usage:  "Validates and runs schema migrations",
		Action: cli.ShowSubcommandHelp,
		Commands: []*cli.Command{
			cliutil.Up(appName, runnerFactory, out, false),
			cliutil.UpTo(appName, runnerFactory, out, false),
			cliutil.DownTo(appName, runnerFactory, out, false),
			cliutil.Validate(appName, runnerFactory, out),
			cliutil.AddLog(appName, runnerFactory, out),
		},
	}

	return command.RunContext(ctx, args)
}

func newRunnerFactory() func(ctx context.Context, schemaNames []string) (cliutil.Runner, error) {
	observationContext := &observation.Context{
		Logger:     sglog.Scoped("runner", ""),
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
