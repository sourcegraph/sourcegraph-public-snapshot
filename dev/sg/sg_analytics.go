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
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
			Description: "Uses OKAYHQ_TOKEN, or fetches a token from gcloud or 1password.",
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

					var errs error
					for _, secret := range []secrets.ExternalSecret{
						{
							Provider: secrets.ExternalProvider1Pass,
							Project:  "Shared",
							Name:     "ttdgfcufz3jggx3d57g6rwodwi",
							Field:    "credential",
						},
						{
							Provider: secrets.ExternalProviderGCloud,
							Project:  "sourcegraph-ci",
							Name:     "CI_OKAYHQ_TOKEN",
						},
					} {
						okayToken, err = store.GetExternal(cmd.Context, secret)
						if err != nil {
							errs = errors.Append(errs, err)
							continue // try the next provider
						}
						if okayToken != "" {
							std.Out.Writef("Got OkayHQ token from %s!", secret.Provider)
							break // done!
						}
					}

					// If we've tried all providers and still don't have the token, we
					// return the error.
					if okayToken == "" {
						return errors.Wrap(errs, "failed to get OkayHQ token")
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

						entry := fmt.Sprintf("- [%s] `%s`", ts, ev.Name)
						if len(ev.Labels) > 0 {
							entry += fmt.Sprintf(": %s", strings.Join(ev.Labels, ", "))
						}
						out.WriteString(entry + fmt.Sprintf(" _(%s)_", strings.Join(metrics, ", ")))

						out.WriteString("\n")
					}
				}

				out.WriteString("\nTo submit these events, use `sg analytics submit`.\n")

				return std.Out.WriteMarkdown(out.String())
			},
		},
	},
}
