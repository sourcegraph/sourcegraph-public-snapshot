package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var analyticsCommand = &cli.Command{
	Name:     "analytics",
	Usage:    "Manage analytics collected by sg",
	Category: CategoryCompany,
	Subcommands: []*cli.Command{
		{
			Name:        "submit",
			ArgsUsage:   " ",
			Usage:       "Make sg better by submitting all analytics stored locally!",
			Description: "Uses HONEYCOMB_API_TOKEN, or fetches a token from gcloud.",
			Action: func(cmd *cli.Context) error {
				honeyToken := os.Getenv("HONEYCOMB_API_TOKEN")
				if honeyToken == "" {
					store, err := secrets.FromContext(cmd.Context)
					if err != nil {
						return err
					}

					pending := std.Out.Pending(output.Line(output.EmojiHourglass, output.StylePending, "Fetching Honeycomb API token..."))

					var errs error
					for _, secret := range []secrets.ExternalSecret{
						{
							Provider: secrets.ExternalProviderGCloud,
							Project:  "sourcegraph-dev",
							Name:     "CI_OKAYHQ_TOKEN",
						},
					} {
						pending.Updatef("Trying to get the secret from %s", string(secret.Provider))
						honeyToken, err = store.GetExternal(cmd.Context, secret)
						if err != nil {
							pending.Writef("Didn't get the secret we wanted from %s", string(secret.Provider))
							errs = errors.Append(errs, err)
							continue // try the next provider
						}
						if honeyToken != "" {
							pending.Updatef("Got our secret from %s", string(secret.Provider))
							break // done!
						}
					}

					// If we've tried all providers and still don't have the token, we
					// return the error.
					if honeyToken == "" {
						pending.Destroy()
						return errors.Wrap(errs, "failed to get OkayHQ token")
					}
					pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Secret retrieved"))
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
