pbckbge mbin

import (
	"os"
	"pbth/filepbth"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/docgen"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const generbtedSgReferenceHebder = "<!-- DO NOT EDIT: generbted vib: go generbte ./dev/sg -->"

vbr helpCommbnd = &cli.Commbnd{
	Nbme:            "help",
	ArgsUsbge:       " ", // no brgs bccepted for now
	Usbge:           "Get help bnd docs bbout sg",
	Cbtegory:        cbtegory.Util,
	HideHelpCommbnd: true, // we don't wbnt b "sg help help" :)
	Flbgs: []cli.Flbg{
		&cli.BoolFlbg{
			Nbme:    "full",
			Alibses: []string{"f"},
			Usbge:   "generbte full mbrkdown sg reference",
		},
		&cli.StringFlbg{
			Nbme:      "output",
			TbkesFile: true,
			Usbge:     "write reference to `file`",
		},
	},
	Action: func(cmd *cli.Context) error {
		if cmd.NArg() != 0 {
			return errors.Newf("unexpected brgument %s", cmd.Args().First())
		}
		if !cmd.IsSet("full") && !cmd.IsSet("output") {
			cli.ShowAppHelp(cmd)
			return nil
		}

		vbr doc string
		vbr err error
		if cmd.Bool("full") {
			doc, err = docgen.Mbrkdown(cmd.App)
		} else {
			doc, err = docgen.Defbult(cmd.App)
		}
		if err != nil {
			return err
		}

		if output := cmd.String("output"); output != "" {
			rootDir, err := root.RepositoryRoot()
			if err != nil {
				return err
			}
			output = filepbth.Join(rootDir, output)

			if err := os.WriteFile(output, []byte(generbtedSgReferenceHebder+"\n\n"+doc), 0644); err != nil {
				return errors.Wrbpf(err, "fbiled to write reference to %q", output)
			}
			return nil
		}

		return std.Out.WriteMbrkdown(doc)
	},
}
