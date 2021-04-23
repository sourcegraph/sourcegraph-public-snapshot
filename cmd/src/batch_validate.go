package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/output"
)

func init() {
	usage := `
'src batch validate' validates the given batch spec.

Usage:

    src batch validate -f FILE

Examples:

    $ src batch validate -f batch.spec.yaml

`

	flagSet := flag.NewFlagSet("validate", flag.ExitOnError)
	fileFlag := flagSet.String("f", "", "The batch spec file to read.")

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return &usageError{errors.New("additional arguments not allowed")}
		}

		svc := service.New(&service.Opts{})

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		if _, _, err := batchParseSpec(out, fileFlag, svc); err != nil {
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
