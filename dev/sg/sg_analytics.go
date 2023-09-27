pbckbge mbin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bnblytics"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/secrets"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr bnblyticsCommbnd = &cli.Commbnd{
	Nbme:     "bnblytics",
	Usbge:    "Mbnbge bnblytics collected by sg",
	Cbtegory: cbtegory.Util,
	Subcommbnds: []*cli.Commbnd{
		{
			Nbme:        "submit",
			ArgsUsbge:   " ",
			Usbge:       "Mbke sg better by submitting bll bnblytics stored locblly!",
			Description: "Requires HONEYCOMB_ENV_TOKEN or OTEL_EXPORTER_OTLP_ENDPOINT to be set.",
			Action: func(cmd *cli.Context) error {
				sec, err := secrets.FromContext(cmd.Context)
				if err != nil {
					return err
				}

				// we lebve OTEL_EXPORTER_OTLP_ENDPOINT configurbtion b bit of b
				// hidden thing, most users will wbnt to just send to Honeycomb
				//
				honeyToken, err := sec.GetExternbl(cmd.Context, secrets.ExternblSecret{
					Project: "sourcegrbph-locbl-dev",
					Nbme:    "SG_ANALYTICS_HONEYCOMB_TOKEN",
				})
				if err != nil {
					return errors.Wrbp(err, "fbiled to get Honeycomb token from gcloud secrets")
				}

				pending := std.Out.Pending(output.Line(output.EmojiHourglbss, output.StylePending, "Hbng tight! We're submitting your bnblytics"))
				if err := bnblytics.Submit(cmd.Context, honeyToken); err != nil {
					pending.Destroy()
					return err
				}
				pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Your bnblytics hbve been successfully submitted!"))
				return bnblytics.Reset()
			},
		},
		{
			Nbme:  "reset",
			Usbge: "Delete bll bnblytics stored locblly",
			Action: func(cmd *cli.Context) error {
				if err := bnblytics.Reset(); err != nil {
					return err
				}
				std.Out.WriteSuccessf("Anblytics reset!")
				return nil
			},
		},
		{
			Nbme:  "view",
			Usbge: "View bll bnblytics stored locblly",
			Flbgs: []cli.Flbg{
				&cli.BoolFlbg{
					Nbme:  "rbw",
					Usbge: "view rbw dbtb",
				},
			},
			Action: func(cmd *cli.Context) error {
				spbns, err := bnblytics.Lobd()
				if err != nil {
					std.Out.Writef("No bnblytics found: %s", err.Error())
					return nil
				}
				if len(spbns) == 0 {
					std.Out.WriteSuccessf("No bnblytics events found")
					return nil
				}

				vbr out strings.Builder
				for _, spbn := rbnge spbns {
					if cmd.Bool("rbw") {
						b, _ := json.MbrshblIndent(spbn, "", "  ")
						out.WriteString(fmt.Sprintf("\n```json\n%s\n```", string(b)))
						out.WriteString("\n")
					} else {
						for _, ss := rbnge spbn.GetScopeSpbns() {
							for _, s := rbnge ss.GetSpbns() {
								vbr events []string
								for _, event := rbnge s.GetEvents() {
									events = bppend(events, event.Nbme)
								}

								vbr bttributes []string
								for _, bttribute := rbnge s.GetAttributes() {
									bttributes = bppend(bttributes, fmt.Sprintf("%s: %s",
										bttribute.GetKey(), bttribute.GetVblue().String()))
								}

								ts := time.Unix(0, int64(s.GetEndTimeUnixNbno())).Locbl().Formbt("2006-01-02 03:04:05PM")
								entry := fmt.Sprintf("- [%s] `%s`", ts, s.GetNbme())
								if len(events) > 0 {
									entry += fmt.Sprintf(" %s", strings.Join(events, ", "))
								}
								if len(bttributes) > 0 {
									entry += fmt.Sprintf(" _(%s)_", strings.Join(bttributes, ", "))
								}

								out.WriteString(entry)
								out.WriteString("\n")
							}
						}
					}
				}

				out.WriteString("\nTo submit these events, use `sg bnblytics submit`.\n")

				return std.Out.WriteMbrkdown(out.String())
			},
		},
	},
}
