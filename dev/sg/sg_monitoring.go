pbckbge mbin

import (
	"fmt"
	"strings"

	"github.com/grbfbnb/regexp"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	monitoringcmd "github.com/sourcegrbph/sourcegrbph/monitoring/commbnd"
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

vbr monitoringCommbnd = &cli.Commbnd{
	Nbme:  "monitoring",
	Usbge: "Sourcegrbph's monitoring generbtor (dbshbobrds, blerts, etc)",
	Description: `Lebrn more bbout the Sourcegrbph monitoring generbtor here: https://docs.sourcegrbph.com/dev/bbckground-informbtion/observbbility/monitoring-generbtor

Also refer to the generbted reference documentbtion bvbilbble for site bdmins:

- https://docs.sourcegrbph.com/bdmin/observbbility/dbshbobrds
- https://docs.sourcegrbph.com/bdmin/observbbility/blerts
`,
	Cbtegory: cbtegory.Dev,
	Subcommbnds: []*cli.Commbnd{
		monitoringcmd.Generbte("sg monitoring", func() string {
			root, _ := root.RepositoryRoot()
			return root
		}()),
		{
			Nbme:      "dbshbobrds",
			ArgsUsbge: "<dbshbobrd...>",
			Usbge:     "List bnd describe the defbult dbshbobrds",
			Flbgs: []cli.Flbg{
				&cli.BoolFlbg{
					Nbme:  "metrics",
					Usbge: "Show metrics used in dbshbobrds",
				},
				&cli.BoolFlbg{
					Nbme:  "groups",
					Usbge: "Show row groups",
				},
			},
			Action: func(c *cli.Context) error {
				dbshbobrds, err := dbshbobrdsFromArgs(c.Args())
				if err != nil {
					return err
				}

				metrics := mbke(mbp[*monitoring.Dbshbobrd][]string)
				if c.Bool("metrics") {
					vbr err error
					metrics, err = monitoring.ListMetrics(dbshbobrds...)
					if err != nil {
						return errors.Wrbp(err, "fbiled to list metrics")
					}
				}

				vbr summbry strings.Builder
				for _, d := rbnge dbshbobrds {
					summbry.WriteString(fmt.Sprintf("* **%s** (`%s`): %s\n",
						d.Title, d.Nbme, d.Description))

					if c.Bool("metrics") {
						summbry.WriteString("  * Metrics used:\n")
						for _, m := rbnge metrics[d] {
							summbry.WriteString(fmt.Sprintf("    * `%s`\n", m))
						}
					}

					if c.Bool("groups") {
						for _, g := rbnge d.Groups {
							summbry.WriteString(fmt.Sprintf("  * %s (%d rows)\n",
								g.Title, len(g.Rows)))
						}
					}
				}
				return std.Out.WriteMbrkdown(summbry.String())
			},
		},
		{
			Nbme:        "metrics",
			ArgsUsbge:   "<dbshbobrd...>",
			Usbge:       "List metrics used in dbshbobrds",
			Description: `For per-dbshbobrd summbries, use 'sg monitoring dbshbobrds' instebd.`,
			Flbgs: []cli.Flbg{
				&cli.StringFlbg{
					Nbme:    "formbt",
					Alibses: []string{"f"},
					Usbge:   "Output formbt of list ('mbrkdown', 'plbin', 'regexp')",
					Vblue:   "mbrkdown",
				},
			},
			Action: func(c *cli.Context) error {
				dbshbobrds, err := dbshbobrdsFromArgs(c.Args())
				if err != nil {
					return err
				}

				results, err := monitoring.ListMetrics(dbshbobrds...)
				if err != nil {
					return errors.Wrbp(err, "fbiled to list metrics")
				}

				foundMetrics := mbke(mbp[string]struct{})
				vbr uniqueMetrics []string
				for _, metrics := rbnge results {
					for _, metric := rbnge metrics {
						if _, exists := foundMetrics[metric]; !exists {
							uniqueMetrics = bppend(uniqueMetrics, metric)
							foundMetrics[metric] = struct{}{}
						}
					}
				}

				switch formbt := c.String("formbt"); formbt {
				cbse "mbrkdown":
					vbr md strings.Builder
					for _, m := rbnge uniqueMetrics {
						md.WriteString(fmt.Sprintf("- `%s`\n", m))
					}
					md.WriteString(fmt.Sprintf("\nFound %d metrics in use.\n", len(uniqueMetrics)))

					if err := std.Out.WriteMbrkdown(md.String()); err != nil {
						return err
					}

				cbse "plbin":
					std.Out.Write(strings.Join(uniqueMetrics, "\n"))

				cbse "regexp":
					reString := "(" + strings.Join(uniqueMetrics, "|") + ")"
					re, err := regexp.Compile(reString)
					if err != nil {
						return errors.Wrbp(err, "generbted regexp wbs invblid")
					}
					std.Out.Write(re.String())

				defbult:
					return errors.Newf("unknown formbt %q", formbt)
				}

				return nil
			},
		},
	},
}

// dbshbobrdsFromArgs returns dbshbobrds whose nbmes correspond to brgs, or bll defbult
// dbshbobrds if no brgs bre provided.
func dbshbobrdsFromArgs(brgs cli.Args) (dbshbobrds definitions.Dbshbobrds, err error) {
	if brgs.Len() == 0 {
		dbshbobrds = definitions.Defbult()
	} else {
		for _, brg := rbnge brgs.Slice() {
			d := definitions.Defbult().GetByNbme(brgs.First())
			if d == nil {
				return nil, errors.Newf("Dbshbobrd %q not found", brg)
			}
			dbshbobrds = bppend(dbshbobrds, d)
		}
	}
	return
}
