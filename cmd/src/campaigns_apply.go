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
'src campaigns apply' is used to apply a campaign spec on a Sourcegraph
instance, creating or updating the described campaign if necessary.

Usage:

    src campaigns apply -f FILE -namespace NAMESPACE [command options]

Examples:

    $ src campaigns apply -f campaign.spec.yaml -namespace myorg

`

	flagSet := flag.NewFlagSet("apply", flag.ExitOnError)
	flags := newCampaignsApplyFlags(flagSet, campaignsDefaultCacheDir())

	doApply := func(ctx context.Context, out *output.Output, svc *campaigns.Service, flags *campaignsApplyFlags) error {
		id, _, err := campaignsExecute(ctx, out, svc, flags)
		if err != nil {
			return err
		}

		pending := campaignsCreatePending(out, "Applying campaign spec")
		campaign, err := svc.ApplyCampaign(ctx, id)
		if err != nil {
			return err
		}
		campaignsCompletePending(pending, "Applying campaign spec")

		out.Write("")
		block := out.Block(output.Line(campaignsSuccessEmoji, campaignsSuccessColor, "Campaign applied!"))
		defer block.Close()

		block.Write("To view the campaign, go to:")
		block.Writef("%s%s", cfg.Endpoint, campaign.URL)

		return nil
	}

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

		if err := doApply(ctx, out, svc, flags); err != nil {
			out.Write("")
			block := out.Block(output.Line("❌", output.StyleWarning, "Error"))
			block.Write(err.Error())
			block.Close()
			out.Write("")
		}

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
