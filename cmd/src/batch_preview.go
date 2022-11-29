package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
'src batch preview' executes the steps in a batch spec and uploads it to a
Sourcegraph instance, ready to be previewed and applied.

Usage:

    src batch preview [command options] [-f FILE]
    src batch preview [command options] FILE

Examples:

    $ src batch preview -f batch.spec.yaml

    $ src batch preview batch.spec.yaml

`

	flagSet := flag.NewFlagSet("preview", flag.ExitOnError)
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

			// Do not apply the uploaded batch spec
			applyBatchSpec: false,
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
