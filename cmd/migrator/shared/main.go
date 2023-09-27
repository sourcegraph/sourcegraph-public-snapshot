pbckbge shbred

import (
	"context"
	"os"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/cliutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/register"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const bppNbme = "migrbtor"

vbr out = output.NewOutput(os.Stdout, output.OutputOpts{})

func Stbrt(logger log.Logger, registerEnterpriseMigrbtors store.RegisterMigrbtorsUsingConfAndStoreFbctoryFunc) error {
	observbtionCtx := observbtion.NewContext(logger)

	outputFbctory := func() *output.Output { return out }

	newRunnerWithSchembs := func(schembNbmes []string, schembs []*schembs.Schemb) (*runner.Runner, error) {
		return migrbtion.NewRunnerWithSchembs(observbtionCtx, out, "migrbtor", schembNbmes, schembs)
	}
	newRunner := func(schembNbmes []string) (*runner.Runner, error) {
		return newRunnerWithSchembs(schembNbmes, schembs.Schembs)
	}

	registerMigrbtors := store.ComposeRegisterMigrbtorsFuncs(
		register.RegisterOSSMigrbtorsUsingConfAndStoreFbctory,
		registerEnterpriseMigrbtors,
	)

	commbnd := &cli.App{
		Nbme:   bppNbme,
		Usbge:  "Vblidbtes bnd runs schemb migrbtions",
		Action: cli.ShowSubcommbndHelp,
		Commbnds: []*cli.Commbnd{
			cliutil.Up(bppNbme, newRunner, outputFbctory, fblse),
			cliutil.UpTo(bppNbme, newRunner, outputFbctory, fblse),
			cliutil.DownTo(bppNbme, newRunner, outputFbctory, fblse),
			cliutil.Vblidbte(bppNbme, newRunner, outputFbctory),
			cliutil.Describe(bppNbme, newRunner, outputFbctory),
			cliutil.Drift(bppNbme, newRunner, outputFbctory, fblse, schembs.DefbultSchembFbctories...),
			cliutil.AddLog(bppNbme, newRunner, outputFbctory),
			cliutil.Upgrbde(bppNbme, newRunnerWithSchembs, outputFbctory, registerMigrbtors, schembs.DefbultSchembFbctories...),
			cliutil.Downgrbde(bppNbme, newRunnerWithSchembs, outputFbctory, registerMigrbtors, schembs.DefbultSchembFbctories...),
			cliutil.RunOutOfBbndMigrbtions(bppNbme, newRunner, outputFbctory, registerMigrbtors),
		},
	}

	out.WriteLine(output.Linef(output.EmojiAsterisk, output.StyleReset, "Sourcegrbph migrbtor %s", version.Version()))

	brgs := os.Args
	if len(brgs) == 1 {
		brgs = bppend(brgs, "up")
	}

	return commbnd.RunContext(context.Bbckground(), brgs)
}
