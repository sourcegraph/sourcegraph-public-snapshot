package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	descriptions "github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	outputFactory := func() *output.Output { return out }
	expectedSchemaFactory := func(filename, version string) (descriptions.SchemaDescription, error) {
		if !regexp.MustCompile(`(^v\d+\.\d+\.\d+$)|(^[A-Fa-f0-9]{40}$)`).MatchString(version) {
			return descriptions.SchemaDescription{}, errors.Newf("failed to parse %q - expected a version of the form `vX.Y.Z` or a 40-character commit hash", version)
		}

		resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/sourcegraph/sourcegraph/%s/%s", version, filename))
		if err != nil {
			return descriptions.SchemaDescription{}, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return descriptions.SchemaDescription{}, errors.Newf("unexpected status %d from github", resp.StatusCode)
		}

		var schemaDescription descriptions.SchemaDescription
		if err := json.NewDecoder(resp.Body).Decode(&schemaDescription); err != nil {
			return descriptions.SchemaDescription{}, err
		}

		return schemaDescription, nil
	}

	command := &cli.App{
		Name:   appName,
		Usage:  "Validates and runs schema migrations",
		Action: cli.ShowSubcommandHelp,
		Commands: []*cli.Command{
			cliutil.Up(appName, runnerFactory, outputFactory, false),
			cliutil.UpTo(appName, runnerFactory, outputFactory, false),
			cliutil.DownTo(appName, runnerFactory, outputFactory, false),
			cliutil.Validate(appName, runnerFactory, outputFactory),
			cliutil.Describe(appName, runnerFactory, outputFactory),
			cliutil.Drift(appName, runnerFactory, outputFactory, expectedSchemaFactory),
			cliutil.AddLog(appName, runnerFactory, outputFactory),
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
