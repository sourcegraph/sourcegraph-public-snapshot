package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/campaigns"
	"github.com/sourcegraph/src-cli/internal/output"
)

func init() {
	usage := `
'src campaigns preview' executes the steps in a campaign spec and uploads it to
a Sourcegraph instance, ready to be previewed and applied.

Usage:

    src campaigns preview -f FILE [command options]

Examples:

    $ src campaigns preview -f campaign.spec.yaml

`

	flagSet := flag.NewFlagSet("preview", flag.ExitOnError)
	flags := newCampaignsApplyFlags(flagSet, campaignsDefaultCacheDir(), campaignsDefaultTempDirPrefix())

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

		svc := campaigns.NewService(&campaigns.ServiceOpts{
			AllowUnsupported: flags.allowUnsupported,
			Client:           cfg.apiClient(flags.api, flagSet.Output()),
		})

		if err := svc.DetermineFeatureFlags(ctx); err != nil {
			return err
		}

		_, url, err := campaignsExecute(ctx, out, svc, flags)
		if err != nil {
			printExecutionError(out, err)
			return &exitCodeError{nil, 1}
		}

		out.Write("")
		block := out.Block(output.Line(campaignsSuccessEmoji, campaignsSuccessColor, "To preview or apply the campaign spec, go to:"))
		defer block.Close()
		block.Writef("%s%s", cfg.Endpoint, url)

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
