package cloud

import "github.com/urfave/cli/v2"

var EphemeralCommand = cli.Command{
	Name:        "ephemeral",
	Aliases:     []string{"eph"},
	Usage:       "Set of commands that operate on Cloud Ephemeral instances",
	Description: "Commands to create, inspect or upgrade Cloud Ephemeral instances",
	Subcommands: []*cli.Command{
		&buildEphemeralCommand,
		&dashboardEphemeralCommand,
		&deleteEphemeralCommand,
		&deployEphemeralCommand,
		&faqEphemeralCommand,
		&leaseEphemeralCommand,
		&listEphemeralCommand,
		&listVersionsEphemeralCommand,
		&opsDashboardEphemeralCommand,
		&statusEphemeralCommand,
		&upgradeEphemeralCommand,
	},
}
