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
	Action:      listCloudEphemeral,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "json",
			Usage: "print instances in JSON format",
		},
	},
}

func listCloudEphemeral(ctx *cli.Context) error {
	// while we work on this command we print a notice and ask to continue
	if err := printWIPNotice(ctx); err != nil {
		return err
	}
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		return err
	}

	cloudClient, err := NewClient(ctx.Context, email, APIEndpoint)
	if err != nil {
		return err
	}

	instances, err := cloudClient.ListInstances(ctx.Context)
	if err != nil {
		return errors.Wrapf(err, "failed to list instances %v", err)
	}
	var printer Printer
	if ctx.Bool("json") {
		printer = &jsonInstancePrinter{w: os.Stdout}
	} else {
		valueFunc := func(inst *Instance) []any {
			name := inst.Name
			if len(name) > 20 {
				name = name[:20]
			}

			status := inst.Status
			createdAt := inst.CreatedAt.String()

			return []any{
				name, status, createdAt,
			}

		}
		printer = newTerminalInstancePrinter(valueFunc, "%-20s %-11s %s", "Name", "Status", "Created At")
	}

	return printer.Print(instances)
}
