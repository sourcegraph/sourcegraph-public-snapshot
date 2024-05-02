package cloud

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ListEphemeralCommand = cli.Command{
	Name:        "list",
	Usage:       "sg could list",
	Description: "list ephemeral cloud instances attached to your GCP account",
	Action:      wipAction(listCloudEphemeral),
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
			Usage: "list all instances, not just those that attached to your GCP account",
		},
	},
}

func listCloudEphemeral(ctx *cli.Context) error {
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		return err
	}

	cloudClient, err := NewClient(ctx.Context, email, APIEndpoint)
	if err != nil {
		return err
	}

	instances, err := cloudClient.ListInstances(ctx.Context, ctx.Bool("all"))
	if err != nil {
		return errors.Wrapf(err, "failed to list instances %v", err)
	}
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
