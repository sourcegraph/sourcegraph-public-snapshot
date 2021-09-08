package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/batches/ui"
	"github.com/sourcegraph/src-cli/internal/cmderrors"

	"github.com/sourcegraph/sourcegraph/lib/output"
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
	flags := newBatchExecuteFlags(flagSet, false, batchDefaultCacheDir(), batchDefaultTempDirPrefix())

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return cmderrors.Usage("additional arguments not allowed")
		}

		ctx, cancel := contextCancelOnInterrupt(context.Background())
		defer cancel()

		var execUI ui.ExecUI
		if flags.textOnly {
			execUI = &ui.JSONLines{}
		} else {
			out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
			execUI = &ui.TUI{Out: out}
		}

		err := executeBatchSpec(ctx, executeBatchSpecOpts{
			flags:  flags,
			client: cfg.apiClient(flags.api, flagSet.Output()),

			// Do not apply the uploaded batch spec
			applyBatchSpec: false,

			ui: execUI,
		})
		if err != nil {
			return cmderrors.ExitCode(1, nil)
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
