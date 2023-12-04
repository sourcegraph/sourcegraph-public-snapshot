package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var analyticsCommand = &cli.Command{
	Name:     "analytics",
	Usage:    "Manage analytics collected by sg",
	Category: category.Util,
	Subcommands: []*cli.Command{
		{
			Name:        "submit",
			ArgsUsage:   " ",
			Usage:       "Make sg better by submitting all analytics stored locally!",
			Description: "Requires HONEYCOMB_ENV_TOKEN or OTEL_EXPORTER_OTLP_ENDPOINT to be set.",
			Action: func(cmd *cli.Context) error {
				sec, err := secrets.FromContext(cmd.Context)
				if err != nil {
					return err
				}

				// we leave OTEL_EXPORTER_OTLP_ENDPOINT configuration a bit of a
				// hidden thing, most users will want to just send to Honeycomb
				//
				honeyToken, err := sec.GetExternal(cmd.Context, secrets.ExternalSecret{
					Project: "sourcegraph-local-dev",
					Name:    "SG_ANALYTICS_HONEYCOMB_TOKEN",
				})
				if err != nil {
					return errors.Wrap(err, "failed to get Honeycomb token from gcloud secrets")
				}

				pending := std.Out.Pending(output.Line(output.EmojiHourglass, output.StylePending, "Hang tight! We're submitting your analytics"))
				if err := analytics.Submit(cmd.Context, honeyToken); err != nil {
					pending.Destroy()
					return err
				}
				pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Your analytics have been successfully submitted!"))
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
				spans, err := analytics.Load()
				if err != nil {
					std.Out.Writef("No analytics found: %s", err.Error())
					return nil
				}
				if len(spans) == 0 {
					std.Out.WriteSuccessf("No analytics events found")
					return nil
				}

				var out strings.Builder
				for _, span := range spans {
					if cmd.Bool("raw") {
						b, _ := json.MarshalIndent(span, "", "  ")
						out.WriteString(fmt.Sprintf("\n```json\n%s\n```", string(b)))
						out.WriteString("\n")
					} else {
						for _, ss := range span.GetScopeSpans() {
							for _, s := range ss.GetSpans() {
								var events []string
								for _, event := range s.GetEvents() {
									events = append(events, event.Name)
								}

								var attributes []string
								for _, attribute := range s.GetAttributes() {
									attributes = append(attributes, fmt.Sprintf("%s: %s",
										attribute.GetKey(), attribute.GetValue().String()))
								}

								ts := time.Unix(0, int64(s.GetEndTimeUnixNano())).Local().Format("2006-01-02 03:04:05PM")
								entry := fmt.Sprintf("- [%s] `%s`", ts, s.GetName())
								if len(events) > 0 {
									entry += fmt.Sprintf(" %s", strings.Join(events, ", "))
								}
								if len(attributes) > 0 {
									entry += fmt.Sprintf(" _(%s)_", strings.Join(attributes, ", "))
								}

								out.WriteString(entry)
								out.WriteString("\n")
							}
						}
					}
				}

				out.WriteString("\nTo submit these events, use `sg analytics submit`.\n")

				return std.Out.WriteMarkdown(out.String())
			},
		},
	},
}
