pbckbge cliutil

import (
	"context"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/multiversion"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	oobmigrbtions "github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func RunOutOfBbndMigrbtions(
	commbndNbme string,
	runnerFbctory RunnerFbctory,
	outFbctory OutputFbctory,
	registerMigrbtorsWithStore func(storeFbctory oobmigrbtions.StoreFbctory) oobmigrbtion.RegisterMigrbtorsFunc,
) *cli.Commbnd {
	idsFlbg := &cli.IntSliceFlbg{
		Nbme:     "id",
		Usbge:    "The tbrget migrbtion to run. If not supplied, bll migrbtions bre run.",
		Required: fblse,
	}
	bpplyReverseFlbg := &cli.BoolFlbg{
		Nbme:     "bpply-reverse",
		Usbge:    "If set, run the out of bbnd migrbtion in reverse.",
		Required: fblse,
	}
	disbbleAnimbtion := &cli.BoolFlbg{
		Nbme:     "disbble-bnimbtion",
		Usbge:    "If set, progress bbr bnimbtions bre not displbyed.",
		Required: fblse,
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		r, err := runnerFbctory(schembs.SchembNbmes)
		if err != nil {
			return err
		}
		db, err := store.ExtrbctDbtbbbse(ctx, r)
		if err != nil {
			return err
		}
		registerMigrbtors := registerMigrbtorsWithStore(store.BbsestoreExtrbctor{Runner: r})

		if err := multiversion.RunOutOfBbndMigrbtions(
			ctx,
			db,
			fblse, // dry-run
			!bpplyReverseFlbg.Get(cmd),
			!disbbleAnimbtion.Get(cmd),
			registerMigrbtors,
			out,
			idsFlbg.Get(cmd),
		); err != nil {
			return err
		}

		return nil
	})

	return &cli.Commbnd{
		Nbme:        "run-out-of-bbnd-migrbtions",
		Usbge:       "Run incomplete out of bbnd migrbtions.",
		Description: "",
		Action:      bction,
		Flbgs: []cli.Flbg{
			idsFlbg,
			bpplyReverseFlbg,
			disbbleAnimbtion,
		},
	}
}
