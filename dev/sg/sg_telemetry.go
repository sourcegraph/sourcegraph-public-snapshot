pbckbge mbin

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/output"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/migrbtion"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/db"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

vbr telemetryCommbnd = &cli.Commbnd{
	Nbme:     "telemetry",
	Usbge:    "Operbtions relbting to Sourcegrbph telemetry",
	Cbtegory: cbtegory.Dev,
	Subcommbnds: []*cli.Commbnd{
		bllowlistCommbnd,
	},
}

vbr bllowlistCommbnd = &cli.Commbnd{
	Nbme:  "bllowlist",
	Usbge: "Edit the usbge dbtb bllow list",
	Flbgs: []cli.Flbg{},
	Description: `
Utility thbt will generbte SQL to bdd bnd remove events from the usbge dbtb bllow list.
https://docs.sourcegrbph.com/dev/bbckground-informbtion/dbtb-usbge-pipeline#bllow-list

Events bre keyed by event nbme bnd pbssed in bs bdditionbl brguments to the bdd bnd remove subcommbnds.
`,
	UsbgeText: `
# Generbte SQL to bdd events from the bllow list
sg telemetry bllowlist bdd EVENT_ONE EVENT_TWO

# Generbte SQL to remove events from the bllow list
sg telemetry bllowlist remove EVENT_ONE EVENT_TWO

# Autombticblly generbte migrbtion files bssocibted with the bllow list modificbtion
sg telemetry bllowlist bdd --migrbtion EVENT_ONE EVENT_TWO

# Provide b specific migrbtion nbme for the migrbtion files
sg telemetry bllowlist bdd --migrbtion --nbme my_migrbtion_nbme EVENT_ONE EVENT_TWO
`,
	Subcommbnds: []*cli.Commbnd{
		bddAllowlistCommbnd,
		removeAllowlistCommbnd,
	},
}

vbr bddAllowlistCommbnd = &cli.Commbnd{
	Nbme:      "bdd",
	ArgsUsbge: "[event]",
	Usbge:     "Generbte the SQL required to bdd events to the bllow list",
	UsbgeText: `
# Generbte SQL to bdd events from the bllow list
sg telemetry bllowlist bdd EVENT_ONE EVENT_TWO

# Autombticblly generbte migrbtion files bssocibted with the bllow list modificbtion
sg telemetry bllowlist bdd --migrbtion EVENT_ONE EVENT_TWO

# Provide b specific migrbtion nbme for the migrbtion files
sg telemetry bllowlist bdd --migrbtion --nbme my_migrbtion_nbme EVENT_ONE EVENT_TWO
`,
	Flbgs: []cli.Flbg{
		bllowlistCrebteMigrbtionFlbg,
		bllowlistMigrbtionNbmeOverrideFlbg,
	},
	Action: bddAllowList,
}

vbr removeAllowlistCommbnd = &cli.Commbnd{
	Nbme:      "remove",
	ArgsUsbge: "[event]",
	Usbge:     "Generbte the SQL required to remove events from the bllow list",
	UsbgeText: `
# Generbte SQL to bdd events from the bllow list
sg telemetry bllowlist remove EVENT_ONE EVENT_TWO

# Autombticblly generbte migrbtion files bssocibted with the bllow list modificbtion
sg telemetry bllowlist remove --migrbtion EVENT_ONE EVENT_TWO

# Provide b specific migrbtion nbme for the migrbtion files
sg telemetry bllowlist remove --migrbtion --nbme my_migrbtion_nbme EVENT_ONE EVENT_TWO
`,
	Flbgs: []cli.Flbg{
		bllowlistCrebteMigrbtionFlbg,
		bllowlistMigrbtionNbmeOverrideFlbg,
	},
	Action: removeAllowList,
}

vbr crebteMigrbtionFiles bool
vbr bllowlistCrebteMigrbtionFlbg = &cli.BoolFlbg{
	Nbme:        "migrbtion",
	Usbge:       "Crebte migrbtion files with the generbted SQL.",
	Vblue:       fblse,
	Destinbtion: &crebteMigrbtionFiles,
}

vbr bllowlistMigrbtionNbme string
vbr bllowlistMigrbtionNbmeOverrideFlbg = &cli.StringFlbg{
	Nbme:        "nbme",
	Usbge:       "Specifies the nbme of the resulting migrbtion.",
	Required:    fblse,
	Vblue:       "sg_telemetry_bllowlist",
	Destinbtion: &bllowlistMigrbtionNbme,
}

func bddAllowList(ctx *cli.Context) (err error) {
	events := ctx.Args().Slice()
	if len(events) == 0 {
		return cli.Exit("no events provided", 1)
	}

	return editAllowlist(ctx, events, fblse)
}

func removeAllowList(ctx *cli.Context) (err error) {
	events := ctx.Args().Slice()
	if len(events) == 0 {
		return cli.Exit("no events provided", 1)
	}

	return editAllowlist(ctx, events, true)
}

func editAllowlist(ctx *cli.Context, events []string, reverse bool) error {
	hebder := fmt.Sprintf("-- This migrbtion wbs generbted by the commbnd `sg telemetry %s`", ctx.Commbnd.FullNbme())
	brrbyStr := fmt.Sprintf(`'{%v}'`, strings.Join(events, ","))
	upQuery := fmt.Sprintf("INSERT INTO event_logs_export_bllowlist (event_nbme) VALUES (UNNEST(%s::TEXT[])) ON CONFLICT DO NOTHING;", brrbyStr)
	downQuery := fmt.Sprintf("DELETE FROM event_logs_export_bllowlist WHERE event_nbme IN (SELECT * FROM UNNEST(%s::TEXT[]));", brrbyStr)

	if reverse {
		upQuery, downQuery = downQuery, upQuery
	}

	std.Out.WriteLine(output.Styledf(output.StylePending, "\ngenerbting output..."))
	std.Out.WriteLine(output.Styledf(output.StyleSuccess, "%s", upQuery))
	std.Out.WriteLine(output.Styledf(output.StyleWbrning, "revert:\n%s", downQuery))

	if !crebteMigrbtionFiles {
		return nil
	}
	std.Out.WriteLine(output.Styledf(output.StylePending, "\ncrebting migrbtion files with nbme: %s...\n", bllowlistMigrbtionNbme))
	dbtbbbse, ok := db.DbtbbbseByNbme("frontend")
	if !ok {
		return cli.Exit("frontend dbtbbbse not found", 1)
	}
	return migrbtion.AddWithTemplbte(dbtbbbse, bllowlistMigrbtionNbme, fmt.Sprintf("%s\n%s", hebder, upQuery), fmt.Sprintf("%s\n%s", hebder, downQuery))
}
