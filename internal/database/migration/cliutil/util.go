package cliutil

import (
	"context"
	"flag"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// makeAction creates a new migration action function. It is expected that these
// commands accept zero arguments and define their own flags.
func makeAction(
	outFactory func() *output.Output,
	f func(ctx context.Context, cmd *cli.Context, out *output.Output) error,
) func(cmd *cli.Context) error {
	return func(cmd *cli.Context) error {
		if cmd.NArg() != 0 {
			return flagHelp(outFactory(), "too many arguments")
		}

		return f(cmd.Context, cmd, outFactory())
	}
}

// flagHelp returns an error that prints the specified error message with usage text.
func flagHelp(out *output.Output, message string, args ...any) error {
	out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: "+message, args...))
	return flag.ErrHelp
}

// setupRunner initializes and returns the runner associated witht the given schema.
func setupRunner(ctx context.Context, factory RunnerFactory, schemaNames ...string) (Runner, error) {
	runner, err := factory(ctx, schemaNames)
	if err != nil {
		return nil, err
	}

	return runner, nil
}

// setupStore initializes and returns the store associated witht the given schema.
func setupStore(ctx context.Context, factory RunnerFactory, schemaName string) (Runner, Store, error) {
	runner, err := setupRunner(ctx, factory, schemaName)
	if err != nil {
		return nil, nil, err
	}

	store, err := runner.Store(ctx, schemaName)
	if err != nil {
		return nil, nil, err
	}

	return runner, store, nil
}

// sanitizeSchemaNames sanitizies the given string slice from the user.
func sanitizeSchemaNames(schemaNames []string) ([]string, error) {
	if len(schemaNames) == 1 && schemaNames[0] == "" {
		schemaNames = nil
	}

	if len(schemaNames) == 1 && schemaNames[0] == "all" {
		return schemas.SchemaNames, nil
	}

	return schemaNames, nil
}

// parseTargets parses the given strings as integers.
func parseTargets(targets []string) ([]int, error) {
	if len(targets) == 1 && targets[0] == "" {
		targets = nil
	}

	versions := make([]int, 0, len(targets))
	for _, target := range targets {
		version, err := strconv.Atoi(target)
		if err != nil {
			return nil, err
		}

		versions = append(versions, version)
	}

	return versions, nil
}
