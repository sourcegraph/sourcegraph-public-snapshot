package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
'src batch apply' is used to apply a batch spec on a Sourcegraph instance,
creating or updating the described batch change if necessary.

Usage:

    src batch apply [command options] [-f FILE]
    src batch apply [command options] FILE

Examples:

    $ src batch apply -f batch.spec.yaml

    $ src batch apply -f batch.spec.yaml -namespace myorg

    $ src batch apply batch.spec.yaml

`

	flagSet := flag.NewFlagSet("apply", flag.ExitOnError)
	flags := newBatchExecuteFlags(flagSet, batchDefaultCacheDir(), batchDefaultTempDirPrefix())

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		file, err := getBatchSpecFile(flagSet, &flags.file)
		if err != nil {
			return err
		}

		ctx, cancel := contextCancelOnInterrupt(context.Background())
		defer cancel()

		if err = executeBatchSpec(ctx, executeBatchSpecOpts{
			flags:  flags,
			client: cfg.apiClient(flags.api, flagSet.Output()),
			file:   file,

			applyBatchSpec: true,
		}); err != nil {
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
