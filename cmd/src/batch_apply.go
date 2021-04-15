package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/output"
)

func init() {
	usage := `
'src batch apply' is used to apply a batch spec on a Sourcegraph instance,
creating or updating the described batch change if necessary.

Usage:

    src batch apply -f FILE [command options]

Examples:

    $ src batch apply -f batch.spec.yaml
  
    $ src batch apply -f batch.spec.yaml -namespace myorg

`

	flagSet := flag.NewFlagSet("apply", flag.ExitOnError)
	flags := newBatchApplyFlags(flagSet, batchDefaultCacheDir(), batchDefaultTempDirPrefix())

	doApply := func(ctx context.Context, out *output.Output, svc *service.Service, flags *batchApplyFlags) error {
		id, _, err := batchExecute(ctx, out, svc, flags)
		if err != nil {
			return err
		}

		pending := batchCreatePending(out, "Applying batch spec")
		batch, err := svc.ApplyBatchChange(ctx, id)
		if err != nil {
			return err
		}
		batchCompletePending(pending, "Applying batch spec")

		out.Write("")
		block := out.Block(output.Line(batchSuccessEmoji, batchSuccessColor, "Batch change applied!"))
		defer block.Close()

		block.Write("To view the batch change, go to:")
		block.Writef("%s%s", cfg.Endpoint, batch.URL)

		return nil
	}

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

		svc := service.New(&service.Opts{
			AllowUnsupported: flags.allowUnsupported,
			AllowIgnored:     flags.allowIgnored,
			Client:           cfg.apiClient(flags.api, flagSet.Output()),
			Workspace:        flags.workspace,
		})

		if err := svc.DetermineFeatureFlags(ctx); err != nil {
			return err
		}

		if err := doApply(ctx, out, svc, flags); err != nil {
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
