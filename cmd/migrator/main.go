package main

import (
	"context"
	"fmt"
	"os"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
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
	run, err := createRunFunc()
	if err != nil {
		return err
	}

	command := cliutil.Flags(appName, run, out)

	if err := command.Parse(args); err != nil {
		return err
	}

	return command.Run(ctx)
}

func createRunFunc() (cliutil.RunFunc, error) {
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	dsns, err := postgresdsn.DSNsBySchema()
	if err != nil {
		return nil, err
	}

	return runner.NewDefaultRunner(dsns, appName, observationContext).Run, nil
}
