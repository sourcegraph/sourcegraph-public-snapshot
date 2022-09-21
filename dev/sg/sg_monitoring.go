package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	monitoringcmd "github.com/sourcegraph/sourcegraph/monitoring/command"
	"github.com/sourcegraph/sourcegraph/monitoring/definitions"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
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
					Name:  "metrics",
					Usage: "Show metrics used in dashboards",
				},
				&cli.BoolFlag{
					Name:  "groups",
					Usage: "Show row groups",
				},
			},
			Action: func(c *cli.Context) error {
				dashboards := definitions.Default()

				metrics := make(map[*monitoring.Dashboard][]string)
				if c.Bool("metrics") {
					var err error
					metrics, err = monitoring.ListMetrics(dashboards...)
					if err != nil {
						return errors.Wrap(err, "failed to list metrics")
					}
				}

				var summary strings.Builder
				for _, d := range dashboards {
					summary.WriteString(fmt.Sprintf("* **%s** (`%s`): %s\n",
						d.Title, d.Name, d.Description))

					if c.Bool("metrics") {
						summary.WriteString("  * Metrics used:\n")
						for _, m := range metrics[d] {
							summary.WriteString(fmt.Sprintf("    * `%s`\n", m))
						}
					}

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
