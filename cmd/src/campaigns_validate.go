package main

import (
	"flag"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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

		specFile, err := campaignsOpenFileFlag(fileFlag)
		if err != nil {
			return err
		}
		defer specFile.Close()

		svc := campaigns.NewService(&campaigns.ServiceOpts{})
		spec, err := svc.ParseCampaignSpec(specFile)
		if err != nil {
			return errors.Wrap(err, "parsing campaign spec")
		}

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		if err := campaignsValidateSpec(out, spec); err != nil {
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

// campaignsValidateSpec validates the given campaign spec. If the spec has
// validation errors, they are output in a human readable form and an
// exitCodeError is returned.
func campaignsValidateSpec(out *output.Output, spec *campaigns.CampaignSpec) error {
	if err := spec.Validate(); err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			block := out.Block(output.Line("\u274c", output.StyleWarning, "Campaign spec failed validation."))

			for i, err := range merr.Errors {
				block.Writef("%d. %s", i+1, err)
			}

			return &exitCodeError{
				error:    nil,
				exitCode: 2,
			}
		} else {
			// This shouldn't happen; let's just punt and let the normal
			// rendering occur.
			return err
		}
	}

	return nil
}
