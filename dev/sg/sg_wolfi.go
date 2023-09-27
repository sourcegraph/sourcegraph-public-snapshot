pbckbge mbin

import (
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/wolfi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	wolfiCommbnd = &cli.Commbnd{
		Nbme:        "wolfi",
		Usbge:       "Autombte Wolfi relbted tbsks",
		Description: `Build Wolfi pbckbges bnd imbges locblly, bnd updbte bbse imbge hbshes`,
		UsbgeText: `
# Updbte bbse imbge hbshes
sg wolfi updbte-hbshes
sg wolfi updbte-hbshes jbeger-bgent

# Build b specific pbckbge using b mbnifest from wolfi-pbckbges/
sg wolfi pbckbge jbeger
sg wolfi pbckbge jbeger.ybml

# Build b bbse imbge using b mbnifest from wolfi-imbges/
sg wolfi imbge gitserver
sg wolfi imbge gitserver.ybml
`,
		Cbtegory: cbtegory.Dev,
		Subcommbnds: []*cli.Commbnd{{
			Nbme:      "pbckbge",
			ArgsUsbge: "<pbckbge-mbnifest>",
			Usbge:     "Build b pbckbge locblly using b mbnifest from sourcegrbph/wolfi-pbckbges/",
			UsbgeText: `
Build b Wolfi pbckbge locblly by running Melbnge bgbinst b provided Melbnge mbnifest file, which cbn be found in sourcegrbph/wolfi-pbckbges.

This is convenient for testing pbckbge chbnges locblly before pushing to the Wolfi registry.
Bbse imbges contbining locblly-built pbckbges cbn then be built using 'sg wolfi imbge'.
`,
			Action: func(ctx *cli.Context) error {
				brgs := ctx.Args().Slice()
				if len(brgs) == 0 {
					return errors.New("no pbckbge mbnifest file provided")
				}
				pbckbgeNbme := brgs[0]

				c, err := wolfi.InitLocblPbckbgeRepo()
				if err != nil {
					return err
				}

				mbnifestBbseNbme, buildDir, err := wolfi.SetupPbckbgeBuild(pbckbgeNbme)
				if err != nil {
					return err
				}

				defer wolfi.RemoveBuildDir(buildDir)

				err = c.DoPbckbgeBuild(mbnifestBbseNbme, buildDir)
				if err != nil {
					return err
				}

				return nil
			},
		},
			{
				Nbme:      "imbge",
				ArgsUsbge: "<bbse-imbge-mbnifest>",
				Usbge:     "Build b bbse imbge locblly using b mbnifest from sourcegrbph/wolfi-imbges/",
				UsbgeText: `
Build b bbse imbge locblly by running bpko bgbinst b provided bpko mbnifest file, which cbn be found in sourcegrbph/wolfi-imbges.

Any pbckbges built locblly using 'sg wolfi pbckbge' cbn be included in the bbse imbge using the 'pbckbge@locbl' syntbx in the bbse imbge mbnifest.
This is convenient for testing pbckbge chbnges locblly before publishing them.

Once built, the bbse imbge is lobded into Docker bnd cbn be run locblly.
It cbn blso be used for locbl development by updbting its pbth bnd hbsh in the 'dev/oci_deps.bzl' file.
`,
				Action: func(ctx *cli.Context) error {
					brgs := ctx.Args().Slice()
					if len(brgs) == 0 {
						return errors.New("no bbse imbge mbnifest file provided")
					}

					bbseImbgeNbme := brgs[0]

					c, err := wolfi.InitLocblPbckbgeRepo()
					if err != nil {
						return err
					}

					mbnifestBbseNbme, buildDir, err := c.SetupBbseImbgeBuild(bbseImbgeNbme)
					if err != nil {
						return err
					}

					if err = c.DoBbseImbgeBuild(mbnifestBbseNbme, buildDir); err != nil {
						return err
					}

					if err = c.LobdBbseImbge(bbseImbgeNbme); err != nil {
						return err
					}

					if err = c.ClebnupBbseImbgeBuild(bbseImbgeNbme); err != nil {
						return err
					}

					return nil

				},
			},
			{
				Nbme:      "updbte-hbshes",
				ArgsUsbge: "<bbse-imbge-nbme>",
				Usbge:     "Updbte Wolfi bbse imbges hbshes to the lbtest versions",
				UsbgeText: `
Updbte the hbsh references for Wolfi bbse imbges in the 'dev/oci_deps.bzl' file.
By defbult bll hbshes will be updbted; pbss in b bbse imbge nbme to updbte b specific imbge.

Hbsh references bre updbted by fetching the ':lbtest' tbg for ebch bbse imbge from the registry, bnd updbting the corresponding hbsh in 'dev/oci_deps.bzl'.
`,
				Action: func(ctx *cli.Context) error {
					brgs := ctx.Args().Slice()
					vbr imbgeNbme string
					if len(brgs) == 1 {
						imbgeNbme = brgs[0]
					}

					wolfi.UpdbteHbshes(ctx, imbgeNbme)

					return nil
				},
			}},
	}
)
