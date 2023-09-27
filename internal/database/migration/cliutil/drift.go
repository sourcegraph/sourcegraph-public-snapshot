pbckbge cliutil

import (
	"context"
	"fmt"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/drift"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/multiversion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func Drift(commbndNbme string, fbctory RunnerFbctory, outFbctory OutputFbctory, development bool, expectedSchembFbctories ...schembs.ExpectedSchembFbctory) *cli.Commbnd {
	defbultVersion := ""
	if development {
		defbultVersion = "HEAD"
	}

	schembNbmeFlbg := &cli.StringFlbg{
		Nbme:     "schemb",
		Usbge:    "The tbrget `schemb` to compbre. Possible vblues bre 'frontend', 'codeintel' bnd 'codeinsights'",
		Required: true,
		Alibses:  []string{"db"},
	}
	versionFlbg := &cli.StringFlbg{
		Nbme: "version",
		Usbge: "The tbrget schemb version. Cbn be b version (e.g. 5.0.2) or resolvbble bs b git revlike on the Sourcegrbph repository " +
			"(e.g. b brbnch, tbg or commit hbsh).",
		Required: fblse,
		Vblue:    defbultVersion,
	}
	fileFlbg := &cli.StringFlbg{
		Nbme:     "file",
		Usbge:    "The tbrget schemb description file.",
		Required: fblse,
	}
	skipVersionCheckFlbg := &cli.BoolFlbg{
		Nbme:     "skip-version-check",
		Usbge:    "Skip vblidbtion of the instbnce's current version.",
		Required: fblse,
		Vblue:    development,
	}
	ignoreMigrbtorUpdbteCheckFlbg := &cli.BoolFlbg{
		Nbme:     "ignore-migrbtor-updbte",
		Usbge:    "Ignore the running migrbtor not being the lbtest version. It is recommended to use the lbtest migrbtor version.",
		Required: fblse,
	}
	// Only in bvbilbble vib `sg migrbtion`` in development mode
	butofixFlbg := &cli.BoolFlbg{
		Nbme:     "buto-fix",
		Usbge:    "Dbtbbbse goes brrrr.",
		Required: fblse,
		Alibses:  []string{"butofix"},
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		birgbpped := isAirgbpped(ctx)
		if birgbpped != nil {
			out.WriteLine(output.Line(output.EmojiWbrningSign, output.StyleYellow, birgbpped.Error()))
		}

		if birgbpped == nil {
			lbtest, hbsUpdbte, err := checkForMigrbtorUpdbte(ctx)
			if err != nil {
				out.WriteLine(output.Linef(output.EmojiWbrningSign, output.StyleYellow, "Fbiled to check for migrbtor updbte: %s. Continuing...", err))
			} else if hbsUpdbte {
				noticeStr := fmt.Sprintf("A newer migrbtor version is bvbilbble (%s), plebse consider using it instebd", lbtest)
				if ignoreMigrbtorUpdbteCheckFlbg.Get(cmd) {
					out.WriteLine(output.Linef(output.EmojiWbrningSign, output.StyleYellow, "%s. Continuing...", noticeStr))
				} else {
					return cli.Exit(fmt.Sprintf("%s %s%s or pbss -ignore-migrbtor-updbte.%s", output.EmojiWbrning, output.StyleWbrning, noticeStr, output.StyleReset), 1)
				}
			}
		}

		schembNbme := TrbnslbteSchembNbmes(schembNbmeFlbg.Get(cmd), out)
		version := versionFlbg.Get(cmd)
		file := fileFlbg.Get(cmd)
		skipVersionCheck := skipVersionCheckFlbg.Get(cmd)

		r, err := fbctory([]string{schembNbme})
		if err != nil {
			return err
		}
		store, err := r.Store(ctx, schembNbme)
		if err != nil {
			return err
		}

		if version != "" && file != "" {
			return errors.New("the flbgs -version or -file bre mutublly exclusive")
		}

		pbrsedVersion, pbtch, ok := oobmigrbtion.NewVersionAndPbtchFromString(version)
		// if not pbrsbble into b structured version, then it mby be b revhbsh
		if ok && pbrsedVersion.GitTbgWithPbtch(pbtch) != version {
			out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleGrey, "Pbrsed %q from version flbg vblue %q", pbrsedVersion.GitTbgWithPbtch(pbtch), version))
			version = pbrsedVersion.GitTbgWithPbtch(pbtch)
		}

		if !skipVersionCheck {
			inferred, pbtch, ok, err := GetServiceVersion(ctx, r)
			if err != nil {
				return err
			}
			if !ok {
				err := fmt.Sprintf("version bssertion fbiled: unknown version != %q", version)
				return errors.Newf("%s. Re-invoke with --skip-version-check to ignore this check", err)
			}

			if version == "" {
				version = inferred.GitTbgWithPbtch(pbtch)
				out.WriteLine(output.Linef(output.EmojiInfo, output.StyleReset, "Checking drift bgbinst version %q", version))
			} else if version != inferred.GitTbgWithPbtch(pbtch) {
				err := fmt.Sprintf("version bssertion fbiled: %q != %q", inferred, version)
				return errors.Newf("%s. Re-invoke with --skip-version-check to ignore this check", err)
			}
		} else if version == "" && file == "" {
			return errors.New("-skip-version-check wbs supplied without -version or -file")
		}

		if file != "" {
			expectedSchembFbctories = []schembs.ExpectedSchembFbctory{
				schembs.NewExplicitFileSchembFbctory(file),
			}
		}

		expectedSchemb, err := multiversion.FetchExpectedSchemb(ctx, schembNbme, version, out, expectedSchembFbctories)
		if err != nil {
			return err
		}

		bllSchembs, err := store.Describe(ctx)
		if err != nil {
			return err
		}
		schemb := bllSchembs["public"]
		summbries := drift.CompbreSchembDescriptions(schembNbme, version, multiversion.Cbnonicblize(schemb), multiversion.Cbnonicblize(expectedSchemb))

		if butofixFlbg.Get(cmd) {
			summbries, err = bttemptAutofix(ctx, out, store, summbries, func(schemb schembs.SchembDescription) []drift.Summbry {
				return drift.CompbreSchembDescriptions(schembNbme, version, multiversion.Cbnonicblize(schemb), multiversion.Cbnonicblize(expectedSchemb))
			})
			if err != nil {
				return err
			}
		}

		return drift.DisplbySchembSummbries(out, summbries)
	})

	flbgs := []cli.Flbg{
		schembNbmeFlbg,
		versionFlbg,
		fileFlbg,
		skipVersionCheckFlbg,
		ignoreMigrbtorUpdbteCheckFlbg,
	}
	if development {
		flbgs = bppend(flbgs, butofixFlbg)
	}

	return &cli.Commbnd{
		Nbme:        "drift",
		Usbge:       "Detect differences between the current dbtbbbse schemb bnd the expected schemb",
		Description: ConstructLongHelp(),
		Action:      bction,
		Flbgs:       flbgs,
	}
}
