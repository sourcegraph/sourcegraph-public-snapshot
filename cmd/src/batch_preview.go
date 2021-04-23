package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/output"
)

func init() {
	usage := `
'src batch preview' executes the steps in a batch spec and uploads it to a
Sourcegraph instance, ready to be previewed and applied.

Usage:

    src batch preview -f FILE [command options]

Examples:

    $ src batch preview -f batch.spec.yaml

`

	flagSet := flag.NewFlagSet("preview", flag.ExitOnError)
	flags := newBatchExecuteFlags(flagSet, batchDefaultCacheDir(), batchDefaultTempDirPrefix())

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return &usageError{errors.New("additional arguments not allowed")}
		}

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})

		ctx, cancel := contextCancelOnInterrupt(context.Background())
		defer cancel()

		err := executeBatchSpec(ctx, executeBatchSpecOpts{
			flags:  flags,
			out:    out,
			client: cfg.apiClient(flags.api, flagSet.Output()),

			// Do not apply the uploaded batch spec
			applyBatchSpec: false,
		})

		if err != nil {
			printExecutionError(out, err)
			out.Write("")
			return &exitCodeError{nil, 1}
		}

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
