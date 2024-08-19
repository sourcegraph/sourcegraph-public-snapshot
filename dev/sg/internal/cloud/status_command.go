package cloud

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var statusEphemeralCommand = cli.Command{
	Name:        "status",
	Usage:       "Get the status of a ephemeral instance",
	Description: "Get the status of the ephemeral instance for this branch or instance with the provided name",
	Action:      statusCloudEphemeral,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			Usage:       "name of the ephemeral instance",
			DefaultText: "current branch name will be used",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "print the instance details in JSON",
		},
		&cli.BoolFlag{
			Name:  "raw",
			Usage: "print all of the instance details",
		},
	},
}

func statusCloudEphemeral(ctx *cli.Context) error {
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		writeGCloudErrorSuggestion()
		return err
	}

	cloudClient, err := NewClient(ctx.Context, email, APIEndpoint)
	if err != nil {
		return err
	}

	name := ctx.String("name")
	if name == "" {
		currentBranch, err := repo.GetCurrentBranch(ctx.Context)
		if err != nil {
			return errors.Wrap(err, "failed to determine current branch")
		}
		name = currentBranch
	}
	name = sanitizeInstanceName(name)

	pending := std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, "Getting status of ephemeral instance %q", name))
	inst, err := cloudClient.GetInstance(ctx.Context, name)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "getting status of %q failed", name))
		return err
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleBold, "Ephemeral instance %q status retrieved", name))

	var printer Printer

	switch {
	case ctx.Bool("json"):
		printer = newJSONInstancePrinter(os.Stdout)
	case ctx.Bool("raw"):
		printer = newRawInstancePrinter(os.Stdout)
	default:
		printer = newDefaultTerminalInstancePrinter()
	}

	std.Out.Write("Ephemeral instance details:")
	printer.Print(inst)
	return nil
}
