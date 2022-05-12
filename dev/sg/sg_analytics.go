package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

var analyticsCommand = &cli.Command{
	Name:     "analytics",
	Usage:    "Manage analytics collected by sg",
	Category: CategoryCompany,
	Subcommands: []*cli.Command{
		{
			Name:        "submit",
			ArgsUsage:   "[github username]",
			Usage:       "Make sg better by submitting all analytics stored locally!",
			Description: "Uses OKAYHQ_TOKEN, or fetches a token from gcloud.",
			Action: func(cmd *cli.Context) error {
				if cmd.Args().Len() != 1 {
					return cli.ShowSubcommandHelp(cmd)
				}

				okayToken := os.Getenv("OKAYHQ_TOKEN")
				if okayToken == "" {
					store, err := secrets.FromContext(cmd.Context)
					if err != nil {
						return err
					}
					okayToken, err = store.GetExternal(cmd.Context, secrets.ExternalSecret{
						Provider: "gcloud",
						Project:  "sourcegraph-ci",
						Name:     "CI_OKAYHQ_TOKEN",
					})
					if err != nil {
						return err
					}
				}

				if err := analytics.Submit(okayToken, cmd.Args().First()); err != nil {
					return err
				}
				std.Out.WriteSuccessf("Analytics successfully submitted!")
				return analytics.Reset()
			},
		},
		{
			Name:  "reset",
			Usage: "Delete all analytics stored locally",
			Action: func(cmd *cli.Context) error {
				if err := analytics.Reset(); err != nil {
					return err
				}
				std.Out.WriteSuccessf("Analytics reset!")
				return nil
			},
		},
		{
			Name:  "view",
			Usage: "View all analytics stored locally",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "raw",
					Usage: "view raw data",
				},
			},
			Action: func(cmd *cli.Context) error {
				events, err := analytics.Load()
				if err != nil {
					std.Out.WriteSuccessf("No analytics found: %s", err.Error())
					return nil
				}

				var out strings.Builder
				out.WriteString(fmt.Sprintf("%d events found:\n", len(events)))

				for _, ev := range events {
					if cmd.Bool("raw") {
						b, _ := json.MarshalIndent(ev, "", "  ")
						out.WriteString(fmt.Sprintf("\n```json\n%s\n```", string(b)))
						out.WriteString("\n")
					} else {
						ts := ev.Timestamp.Local().Format("2006-01-02 03:04:05PM")
						var metrics []string
						for k, v := range ev.Metrics {
							metrics = append(metrics, fmt.Sprintf("%s: %s", k, v.ValueString()))
						}

						entry := fmt.Sprintf("- [%s] `%s`: %s _(%s)_",
							ts, ev.Name, strings.Join(ev.Labels, ", "), strings.Join(metrics, ", "))
						out.WriteString(entry)

						out.WriteString("\n")
					}
				}

				out.WriteString("\nTo submit these events, use `sg analytics submit`.\n")

				return std.Out.WriteMarkdown(out.String())
			},
		},
	},
}
