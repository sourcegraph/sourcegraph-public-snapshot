package cloud

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var listEphemeralCommand = cli.Command{
	Name:        "list",
	Usage:       "List ephemeral instances assigned to you, or all instances",
	Description: "List ephemeral instances assigned to your GCP email, or all instances that are deployed",
	Action:      listCloudEphemeral,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "json",
			Usage: "print the instance details in JSON",
		},
		&cli.BoolFlag{
			Name:  "raw",
			Usage: "print all of the instance details",
		},
		&cli.BoolFlag{
			Name:  "all",
			Usage: "list all ephemeral instances, not just those that attached to your GCP account",
		},
	},
}

func listCloudEphemeral(ctx *cli.Context) error {
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		writeGCloudErrorSuggestion()
		return err
	}

	cloudClient, err := NewClient(ctx.Context, email, APIEndpoint)
	if err != nil {
		return err
	}

	msg := "Fetching list of instances..."
	if !ctx.Bool("all") {
		msg = fmt.Sprintf("Fetching list of instances attached to your GCP account %q", email)
	}

	pending := std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, msg))
	instances, err := cloudClient.ListInstances(ctx.Context, ctx.Bool("all"))
	if err != nil {
		pending.Complete(output.Linef(CloudEmoji, output.StyleFailure, "failed to list instances: %v", err))
		return errors.Wrapf(err, "failed to list instances %v", err)
	}
	pending.Complete(output.Linef(CloudEmoji, output.StyleSuccess, "Fetched %d instances", len(instances)))
	var printer Printer
	switch {
	case ctx.Bool("json"):
		printer = newJSONInstancePrinter(os.Stdout)
	case ctx.Bool("raw"):
		printer = newRawInstancePrinter(os.Stdout)
	default:
		printer = newDefaultTerminalInstancePrinter()
	}

	return printer.Print(instances...)
}
