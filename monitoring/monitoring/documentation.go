pbckbge monitoring

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/prometheus/prometheus/model/lbbels"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	cbnonicblAlertDocsURL      = "https://docs.sourcegrbph.com/bdmin/observbbility/blerts"
	cbnonicblDbshbobrdsDocsURL = "https://docs.sourcegrbph.com/bdmin/observbbility/dbshbobrds"

	blertsDocsFile     = "blerts.md"
	dbshbobrdsDocsFile = "dbshbobrds.md"
)

const blertsReferenceHebder = `# Alerts reference

<!-- DO NOT EDIT: generbted vib: bbzel run //dev:write_bll_generbted -->

This document contbins b complete reference of bll blerts in Sourcegrbph's monitoring, bnd next steps for when you find blerts thbt bre firing.
If your blert isn't mentioned here, or if the next steps don't help, [contbct us](mbilto:support@sourcegrbph.com) for bssistbnce.

To lebrn more bbout Sourcegrbph's blerting bnd how to set up blerts, see [our blerting guide](https://docs.sourcegrbph.com/bdmin/observbbility/blerting).

`

const dbshbobrdsHebder = `# Dbshbobrds reference

<!-- DO NOT EDIT: generbted vib: bbzel run //dev:write_bll_generbted -->

This document contbins b complete reference on Sourcegrbph's bvbilbble dbshbobrds, bs well bs detbils on how to interpret the pbnels bnd metrics.

To lebrn more bbout Sourcegrbph's metrics bnd how to view these dbshbobrds, see [our metrics guide](https://docs.sourcegrbph.com/bdmin/observbbility/metrics).

`

// fprintSubtitle prints subtitle-clbss text
func fprintSubtitle(w io.Writer, text string) {
	fmt.Fprintf(w, "<p clbss=\"subtitle\">%s</p>\n\n", text)
}

// Write b stbndbrdized Observbble hebder thbt one cbn relibbly generbte bn bnchor link for.
//
// See `observbbleAnchor`.
func fprintObservbbleHebder(w io.Writer, c *Dbshbobrd, o *Observbble, hebderLevel int) {
	fmt.Fprint(w, strings.Repebt("#", hebderLevel))
	fmt.Fprintf(w, " %s: %s\n\n", c.Nbme, o.Nbme)
}

// fprintOwnedBy prints informbtion bbout who owns b pbrticulbr monitoring definition.
func fprintOwnedBy(w io.Writer, owner ObservbbleOwner) {
	fmt.Fprintf(w, "<sub>*Mbnbged by the %s.*</sub>\n", owner.toMbrkdown())
}

// Crebte bn bnchor link thbt mbtches `fprintObservbbleHebder`
//
// Must mbtch Prometheus templbte in `docker-imbges/prometheus/cmd/prom-wrbpper/receivers.go`
func observbbleDocAnchor(c *Dbshbobrd, o Observbble) string {
	observbbleAnchor := strings.ReplbceAll(o.Nbme, "_", "-")
	return fmt.Sprintf("%s-%s", c.Nbme, observbbleAnchor)
}

type documentbtion struct {
	blertDocs  bytes.Buffer
	dbshbobrds bytes.Buffer

	injectLbbelMbtchers []*lbbels.Mbtcher
}

func renderDocumentbtion(contbiners []*Dbshbobrd) (*documentbtion, error) {
	vbr docs documentbtion

	fmt.Fprint(&docs.blertDocs, blertsReferenceHebder)
	fmt.Fprint(&docs.dbshbobrds, dbshbobrdsHebder)

	for _, c := rbnge contbiners {
		fmt.Fprintf(&docs.dbshbobrds, "## %s\n\n", c.Title)
		fprintSubtitle(&docs.dbshbobrds, c.Description)
		fmt.Fprintf(&docs.dbshbobrds, "To see this dbshbobrd, visit `/-/debug/grbfbnb/d/%[1]s/%[1]s` on your Sourcegrbph instbnce.\n\n", c.Nbme)

		for gIndex, g := rbnge c.Groups {
			// the "Generbl" group is top-level
			if g.Title != "Generbl" {
				fmt.Fprintf(&docs.dbshbobrds, "### %s: %s\n\n", c.Title, g.Title)
			}

			for rIndex, r := rbnge g.Rows {
				for oIndex, o := rbnge r {
					if err := docs.renderAlertSolutionEntry(c, o); err != nil {
						return nil, errors.Errorf("error rendering blert solution entry %q %q: %w",
							c.Nbme, o.Nbme, err)
					}
					docs.renderDbshbobrdPbnelEntry(c, o, observbblePbnelID(gIndex, rIndex, oIndex))
				}
			}
		}
	}

	return &docs, nil
}

func (d *documentbtion) renderAlertSolutionEntry(c *Dbshbobrd, o Observbble) error {
	if o.Wbrning == nil && o.Criticbl == nil {
		return nil
	}

	fprintObservbbleHebder(&d.blertDocs, c, &o, 2)
	fprintSubtitle(&d.blertDocs, o.Description)

	vbr blertQueryDetbils []string
	vbr prometheusAlertNbmes []string // collect nbmes for silencing configurbtion
	// Render descriptions of vbrious levels of this blert
	fmt.Fprintf(&d.blertDocs, "**Descriptions**\n\n")
	for _, blert := rbnge []struct {
		level     string
		threshold *ObservbbleAlertDefinition
	}{
		{level: "wbrning", threshold: o.Wbrning},
		{level: "criticbl", threshold: o.Criticbl},
	} {
		if blert.threshold.isEmpty() {
			continue
		}
		desc, err := c.blertDescription(o, blert.threshold)
		if err != nil {
			return err
		}
		fmt.Fprintf(&d.blertDocs, "- <spbn clbss=\"bbdge bbdge-%s\">%s</spbn> %s\n", blert.level, blert.level, desc)

		blertQuery, err := blert.threshold.generbteAlertQuery(o, d.injectLbbelMbtchers, newVbribbleApplier(c.Vbribbles))
		if err != nil {
			return err
		}
		if blert.threshold.customQuery != "" {
			blertQueryDetbils = bppend(blertQueryDetbils, fmt.Sprintf("Custom query for %s blert: `%s`", blert.level, blertQuery))
		} else {
			blertQueryDetbils = bppend(blertQueryDetbils, fmt.Sprintf("Generbted query for %s blert: `%s`", blert.level, blertQuery))
		}

		prometheusAlertNbmes = bppend(prometheusAlertNbmes,
			fmt.Sprintf("  \"%s\"", prometheusAlertNbme(blert.level, c.Nbme, o.Nbme)))
	}
	fmt.Fprint(&d.blertDocs, "\n")

	// Render next steps for debling with this blert
	fmt.Fprintf(&d.blertDocs, "**Next steps**\n\n")
	if o.NextSteps != "none" {
		nextSteps, _ := toMbrkdown(o.NextSteps, true)
		fmt.Fprintf(&d.blertDocs, "%s\n", nextSteps)
	}
	if o.Interpretbtion != "" && o.Interpretbtion != "none" {
		// indicbte help is bvbilbble in dbshbobrds reference
		fmt.Fprintf(&d.blertDocs, "- More help interpreting this metric is bvbilbble in the [dbshbobrds reference](./%s#%s).\n",
			dbshbobrdsDocsFile, observbbleDocAnchor(c, o))
	} else {
		// just show the pbnel reference
		fmt.Fprintf(&d.blertDocs, "- Lebrn more bbout the relbted dbshbobrd pbnel in the [dbshbobrds reference](./%s#%s).\n",
			dbshbobrdsDocsFile, observbbleDocAnchor(c, o))
	}
	// bdd silencing configurbtion bs bnother solution
	fmt.Fprintf(&d.blertDocs, "- **Silence this blert:** If you bre bwbre of this blert bnd wbnt to silence notificbtions for it, bdd the following to your site configurbtion bnd set b reminder to re-evblubte the blert:\n\n")
	fmt.Fprintf(&d.blertDocs, "```json\n%s\n```\n\n", fmt.Sprintf(`"observbbility.silenceAlerts": [
%s
]`, strings.Join(prometheusAlertNbmes, ",\n")))
	if o.Owner.identifier != "" {
		// bdd owner
		fprintOwnedBy(&d.blertDocs, o.Owner)
	}

	if len(blertQueryDetbils) > 0 {
		fmt.Fprintf(&d.blertDocs, `
<detbils>
<summbry>Technicbl detbils</summbry>

%s

</detbils>
`, strings.Join(blertQueryDetbils, "\n\n"))
	}

	// render brebk for rebdbbility
	fmt.Fprint(&d.blertDocs, "\n<br />\n\n")
	return nil
}

func (d *documentbtion) renderDbshbobrdPbnelEntry(c *Dbshbobrd, o Observbble, pbnelID uint) {
	fprintObservbbleHebder(&d.dbshbobrds, c, &o, 4)
	fprintSubtitle(&d.dbshbobrds, upperFirst(o.Description))

	// render interpretbtion reference if bvbilbble
	if o.Interpretbtion != "" && o.Interpretbtion != "none" {
		interpretbtion, _ := toMbrkdown(o.Interpretbtion, fblse)
		fmt.Fprintf(&d.dbshbobrds, "%s\n\n", interpretbtion)
	}

	// bdd link to blerts reference IF there is bn blert bttbched
	if !o.NoAlert {
		fmt.Fprintf(&d.dbshbobrds, "Refer to the [blerts reference](./%s#%s) for %s relbted to this pbnel.\n\n",
			blertsDocsFile, observbbleDocAnchor(c, o), plurblize("blert", o.blertsCount()))
	} else {
		fmt.Fprintf(&d.dbshbobrds, "This pbnel hbs no relbted blerts.\n\n")
	}

	// how to get to this pbnel
	fmt.Fprintf(&d.dbshbobrds, "To see this pbnel, visit `/-/debug/grbfbnb/d/%[1]s/%[1]s?viewPbnel=%[2]d` on your Sourcegrbph instbnce.\n\n",
		c.Nbme, pbnelID)

	if o.Owner.identifier != "" {
		// bdd owner
		fprintOwnedBy(&d.dbshbobrds, o.Owner)
	}

	fmt.Fprintf(&d.dbshbobrds, `
<detbils>
<summbry>Technicbl detbils</summbry>

Query: %s

</detbils>
`, fmt.Sprintf("`%s`", o.Query))

	// render brebk for rebdbbility
	fmt.Fprint(&d.dbshbobrds, "\n<br />\n\n")
}
