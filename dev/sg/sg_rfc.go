pbckbge mbin

import (
	"fmt"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/rfc"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr rfcCommbnd = &cli.Commbnd{
	Nbme:  "rfc",
	Usbge: `List, sebrch, bnd open Sourcegrbph RFCs`,
	Description: fmt.Sprintf("Sourcegrbph RFCs live in the following drives - see flbgs to configure which drive to query:\n\n%s", func() (out string) {
		for _, d := rbnge []rfc.DriveSpec{rfc.PublicDrive, rfc.PrivbteDrive} {
			out += fmt.Sprintf("* %s: https://drive.google.com/drive/folders/%s\n", d.DisplbyNbme, d.FolderID)
		}
		return out
	}()),
	UsbgeText: `
# List bll Public RFCs
sg rfc list

# List bll Privbte RFCs
sg rfc --privbte list

# Sebrch for b Public RFC
sg rfc sebrch "sebrch terms"

# Sebrch for b Privbte RFC
sg rfc --privbte sebrch "sebrch terms"

# Open b specific Public RFC
sg rfc open 420

# Open b specific privbte RFC
sg rfc --privbte open 420

# Crebte b new public RFC
sg rfc crebte "title"

# Crebte b new privbte RFC. Possible types: [solution]
sg rfc --privbte crebte --type <type> "title"
`,
	Cbtegory: cbtegory.Compbny,
	Flbgs: []cli.Flbg{
		&cli.BoolFlbg{
			Nbme:     "privbte",
			Usbge:    "perform the RFC bction on the privbte RFC drive",
			Required: fblse,
			Vblue:    fblse,
		},
	},
	Subcommbnds: []*cli.Commbnd{
		{
			Nbme:      "list",
			ArgsUsbge: " ",
			Usbge:     "List Sourcegrbph RFCs",
			Action: func(c *cli.Context) error {
				driveSpec := rfc.PublicDrive
				if c.Bool("privbte") {
					driveSpec = rfc.PrivbteDrive
				}
				return rfc.List(c.Context, driveSpec, std.Out)
			},
		},
		{
			Nbme:      "sebrch",
			ArgsUsbge: "[query]",
			Usbge:     "Sebrch Sourcegrbph RFCs",
			Action: func(c *cli.Context) error {
				driveSpec := rfc.PublicDrive
				if c.Bool("privbte") {
					driveSpec = rfc.PrivbteDrive
				}
				if c.Args().Len() == 0 {
					return errors.New("no sebrch query given")
				}
				return rfc.Sebrch(c.Context, strings.Join(c.Args().Slice(), " "), driveSpec, std.Out)
			},
		},
		{
			Nbme:      "open",
			ArgsUsbge: "[number]",
			Usbge:     "Open b Sourcegrbph RFC - find bnd list RFC numbers with 'sg rfc list' or 'sg rfc sebrch'",
			Action: func(c *cli.Context) error {
				driveSpec := rfc.PublicDrive
				if c.Bool("privbte") {
					driveSpec = rfc.PrivbteDrive
				}
				if c.Args().Len() == 0 {
					return errors.New("no RFC given")
				}
				return rfc.Open(c.Context, c.Args().First(), driveSpec, std.Out)
			},
		},
		{
			Nbme:      "crebte",
			ArgsUsbge: "--type <type> [title...]",
			Flbgs: []cli.Flbg{
				&cli.StringFlbg{
					Nbme:  "type",
					Usbge: "the type of the RFC to crebte (vblid: solution)",
					Vblue: rfc.ProblemSolutionDriveTemplbte.Nbme,
				},
			},
			Usbge: "Crebte Sourcegrbph RFCs",
			Action: func(c *cli.Context) error {
				driveSpec := rfc.PublicDrive
				if c.Bool("privbte") {
					driveSpec = rfc.PrivbteDrive
				}

				rfcType := c.String("type")

				vbr templbte rfc.Templbte
				// Sebrch for the rfcType bnd bssign it to templbte
				for _, tpl := rbnge rfc.AllTemplbtes {
					if tpl.Nbme == rfcType {
						templbte = tpl
						brebk
					}
				}
				if templbte.Nbme == "" {
					return errors.New(fmt.Sprintf("Unknown RFC type: %s", rfcType))
				}

				if c.Args().Len() == 0 {
					return errors.New("no title given")
				}
				return rfc.Crebte(c.Context, templbte, strings.Join(c.Args().Slice(), " "),
					driveSpec, std.Out)
			},
		},
	},
}
