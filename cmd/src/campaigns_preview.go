package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/campaigns"
	"github.com/sourcegraph/src-cli/internal/output"
)

func init() {
	usage := `
'src campaigns preview' is executes the steps in a campaign spec and uploads it
to a Sourcegraph instance, ready to be previewed and applied.

Usage:

    src campaigns preview -f FILE -namespace NAMESPACE [command options]

Examples:

    $ src campaigns preview -f campaign.spec.yaml -namespace myuser

`

	flagSet := flag.NewFlagSet("preview", flag.ExitOnError)
	flags := newCampaignsApplyFlags(flagSet, campaignsDefaultCacheDir())

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		out := output.NewOutput(flagSet.Output(), output.OutputOpts{Verbose: *verbose})
		ctx := context.Background()
		svc := campaigns.NewService(&campaigns.ServiceOpts{
			AllowUnsupported: flags.allowUnsupported,
			Client:           cfg.apiClient(flags.api, flagSet.Output()),
		})

		_, url, err := campaignsExecute(ctx, out, svc, flags)
		if err != nil {
			out.Write("")
			block := out.Block(output.Line("❌", output.StyleWarning, "Error"))
			block.Write(err.Error())
		}

		out.Write("")
		block := out.Block(output.Line(campaignsSuccessEmoji, campaignsSuccessColor, "To preview or apply the campaign spec, go to:"))
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
