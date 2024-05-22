package main

import (
	"fmt"
	"strings"

	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
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
	Description: `Learn more about the Sourcegraph monitoring generator here: https://docs-legacy.sourcegraph.com/dev/background-information/observability/monitoring-generator

Also refer to the generated reference documentation available for site admins:

- https://docs.sourcegraph.com/admin/observability/dashboards
- https://docs.sourcegraph.com/admin/observability/alerts
`,
	Category: category.Dev,
	Subcommands: []*cli.Command{
		monitoringcmd.Generate("sg monitoring", func() string {
			root, _ := root.RepositoryRoot()
			return root
		}()),
		{
			Name:      "dashboards",
			ArgsUsage: "<dashboard...>",
			Usage:     "List and describe the default dashboards",
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
				dashboards, err := dashboardsFromArgs(c.Args())
				if err != nil {
					return err
				}

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
		{
			Name:        "metrics",
			ArgsUsage:   "<dashboard...>",
			Usage:       "List metrics used in dashboards",
			Description: `For per-dashboard summaries, use 'sg monitoring dashboards' instead.`,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "format",
					Aliases: []string{"f"},
					Usage:   "Output format of list ('markdown', 'plain', 'regexp')",
					Value:   "markdown",
				},
			},
			Action: func(c *cli.Context) error {
				dashboards, err := dashboardsFromArgs(c.Args())
				if err != nil {
					return err
				}

				results, err := monitoring.ListMetrics(dashboards...)
				if err != nil {
					return errors.Wrap(err, "failed to list metrics")
				}

				foundMetrics := make(map[string]struct{})
				var uniqueMetrics []string
				for _, metrics := range results {
					for _, metric := range metrics {
						if _, exists := foundMetrics[metric]; !exists {
							uniqueMetrics = append(uniqueMetrics, metric)
							foundMetrics[metric] = struct{}{}
						}
					}
				}

				switch format := c.String("format"); format {
				case "markdown":
					var md strings.Builder
					for _, m := range uniqueMetrics {
						md.WriteString(fmt.Sprintf("- `%s`\n", m))
					}
					md.WriteString(fmt.Sprintf("\nFound %d metrics in use.\n", len(uniqueMetrics)))

					if err := std.Out.WriteMarkdown(md.String()); err != nil {
						return err
					}

				case "plain":
					std.Out.Write(strings.Join(uniqueMetrics, "\n"))

				case "regexp":
					reString := "(" + strings.Join(uniqueMetrics, "|") + ")"
					re, err := regexp.Compile(reString)
					if err != nil {
						return errors.Wrap(err, "generated regexp was invalid")
					}
					std.Out.Write(re.String())

				default:
					return errors.Newf("unknown format %q", format)
				}

				return nil
			},
		},
	},
}

// dashboardsFromArgs returns dashboards whose names correspond to args, or all default
// dashboards if no args are provided.
func dashboardsFromArgs(args cli.Args) (dashboards definitions.Dashboards, err error) {
	if args.Len() == 0 {
		dashboards = definitions.Default()
	} else {
		for _, arg := range args.Slice() {
			d := definitions.Default().GetByName(arg)
			if d == nil {
				return nil, errors.Newf("Dashboard %q not found", arg)
			}
			dashboards = append(dashboards, d)
		}
	}
	return
}
