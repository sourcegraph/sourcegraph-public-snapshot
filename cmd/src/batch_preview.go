package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/batches"
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
	flags := newBatchApplyFlags(flagSet, batchDefaultCacheDir(), batchDefaultTempDirPrefix())

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

		svc := batches.NewService(&batches.ServiceOpts{
			AllowUnsupported: flags.allowUnsupported,
			AllowIgnored:     flags.allowIgnored,
			Client:           cfg.apiClient(flags.api, flagSet.Output()),
			Workspace:        flags.workspace,
		})

		if err := svc.DetermineFeatureFlags(ctx); err != nil {
			return err
		}

		_, url, err := batchExecute(ctx, out, svc, flags)
		if err != nil {
			printExecutionError(out, err)
			out.Write("")
			return &exitCodeError{nil, 1}
		}

		out.Write("")
		block := out.Block(output.Line(batchSuccessEmoji, batchSuccessColor, "To preview or apply the batch spec, go to:"))
		defer block.Close()
		block.Writef("%s%s", cfg.Endpoint, url)

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
