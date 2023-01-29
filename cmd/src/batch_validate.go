package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/batches/ui"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
'src batch validate' validates the given batch spec.

Usage:

    src batch validate [-f] FILE

Examples:

    $ src batch validate batch.spec.yaml

    $ src batch validate -f batch.spec.yaml

`

	flagSet := flag.NewFlagSet("validate", flag.ExitOnError)
	apiFlags := api.NewFlags(flagSet)
	fileFlag := flagSet.String("f", "", "The batch spec file to read, or - to read from standard input.")

	var (
		allowUnsupported bool
		allowIgnored     bool
	)
	flagSet.BoolVar(
		&allowUnsupported, "allow-unsupported", false,
		"Allow unsupported code hosts.",
	)
	flagSet.BoolVar(
		&allowIgnored, "force-override-ignore", false,
		"Do not ignore repositories that have a .batchignore file.",
	)

	handler := func(args []string) error {
		ctx := context.Background()

		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return cmderrors.Usage("additional arguments not allowed")
		}

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		ui := &ui.TUI{Out: out}
		svc := service.New(&service.Opts{
			Client: cfg.apiClient(apiFlags, flagSet.Output()),
		})

		_, ffs, err := svc.DetermineLicenseAndFeatureFlags(ctx)
		if err != nil {
			return err
		}

		if err := validateSourcegraphVersionConstraint(ctx, ffs); err != nil {
			ui.ExecutionError(err)
			return err
		}

		file, err := getBatchSpecFile(flagSet, fileFlag)
		if err != nil {
			return err
		}

		if _, _, _, err := parseBatchSpec(ctx, file, svc); err != nil {
			ui.ParsingBatchSpecFailure(err)
			return err
		}

		out.WriteLine(output.Line("\u2705", output.StyleSuccess, "Batch spec successfully validated."))
		return nil
	}

	batchCommands = append(batchCommands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src batch %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}
