package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/campaigns"
	"github.com/sourcegraph/src-cli/internal/output"
)

func init() {
	usage := `
'src campaigns validate' validates the given campaign spec.

Usage:

    src campaigns validate -f FILE

Examples:

    $ src campaigns validate -f campaign.spec.yaml

`

	flagSet := flag.NewFlagSet("validate", flag.ExitOnError)
	fileFlag := flagSet.String("f", "", "The campaign spec file to read.")

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return &usageError{errors.New("additional arguments not allowed")}
		}

		specFile, err := campaignsOpenFileFlag(fileFlag)
		if err != nil {
			return err
		}
		defer specFile.Close()

		svc := campaigns.NewService(&campaigns.ServiceOpts{})

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		if _, _, err := campaignsParseSpec(out, svc, specFile); err != nil {
			return err
		}

		out.WriteLine(output.Line("\u2705", output.StyleSuccess, "Campaign spec successfully validated."))
		return nil
	}

	campaignsCommands = append(campaignsCommands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}
