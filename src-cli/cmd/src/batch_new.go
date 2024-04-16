package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
'src batch new' creates a new batch spec YAML, prefilled with all required
fields.

Usage:

    src batch new [-f FILE]

Examples:


    $ src batch new -f batch.spec.yaml

`

	flagSet := flag.NewFlagSet("new", flag.ExitOnError)
	apiFlags := api.NewFlags(flagSet)

	var (
		fileFlag = flagSet.String("f", "batch.yaml", "The name of the batch spec file to create.")
	)

	handler := func(args []string) error {
		ctx := context.Background()

		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return cmderrors.Usage("additional arguments not allowed")
		}

		svc := service.New(&service.Opts{
			Client: cfg.apiClient(apiFlags, flagSet.Output()),
		})

		_, ffs, err := svc.DetermineLicenseAndFeatureFlags(ctx)
		if err != nil {
			return err
		}

		if err := validateSourcegraphVersionConstraint(ctx, ffs); err != nil {
			return err
		}

		if err := svc.GenerateExampleSpec(ctx, *fileFlag); err != nil {
			return err
		}

		fmt.Printf("%s created.\n", *fileFlag)
		return nil
	}

	batchCommands = append(batchCommands, &command{
		flagSet: flagSet,
		aliases: []string{},
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src batch %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}
