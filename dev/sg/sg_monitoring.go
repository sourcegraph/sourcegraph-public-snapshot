package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	monitoringcmd "github.com/sourcegraph/sourcegraph/monitoring/command"
	"github.com/sourcegraph/sourcegraph/monitoring/definitions"
)

var monitoringCommand = &cli.Command{
	Name:  "monitoring",
	Usage: "Sourcegraph's monitoring generator (dashboards, alerts, etc)",
	Description: `Learn more about the Sourcegraph monitoring generator here: https://docs.sourcegraph.com/dev/background-information/observability/monitoring-generator

Also refer to the generated reference documentation available for site admins:

- https://docs.sourcegraph.com/admin/observability/dashboards
- https://docs.sourcegraph.com/admin/observability/alerts
`,
	Category: CategoryDev,
	Subcommands: []*cli.Command{
		monitoringcmd.Generate("sg monitoring", func() string {
			root, _ := root.RepositoryRoot()
			return root
		}()),
		{
			Name:  "dashboards",
			Usage: "List and describe the default dashboards",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "groups",
					Usage: "Show row groups",
					Value: true,
				},
			},
			Action: func(c *cli.Context) error {
				dashboards := definitions.Default()
				var summary strings.Builder
				for _, d := range dashboards {
					summary.WriteString(fmt.Sprintf("* **%s** (`%s`): %s\n",
						d.Title, d.Name, d.Description))

					if c.Bool("groups") {
						for _, g := range d.Groups {
							summary.WriteString(fmt.Sprintf("  * %s (%d rows)\n",
								g.Title, len(g.Rows)))
						}
					}
				}
				return std.Out.WriteMarkdown(summary.String())
			},
		},
	},
}
