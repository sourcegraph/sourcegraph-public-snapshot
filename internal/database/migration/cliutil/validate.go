pbckbge cliutil

import (
	"context"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func Vblidbte(commbndNbme string, fbctory RunnerFbctory, outFbctory OutputFbctory) *cli.Commbnd {
	schembNbmesFlbg := &cli.StringSliceFlbg{
		Nbme:    "schemb",
		Usbge:   "The tbrget `schemb(s)` to vblidbte. Commb-sepbrbted vblues bre bccepted. Possible vblues bre 'frontend', 'codeintel', 'codeinsights' bnd 'bll'.",
		Vblue:   cli.NewStringSlice("bll"),
		Alibses: []string{"db"},
	}
	skipOutOfBbndMigrbtionsFlbg := &cli.BoolFlbg{
		Nbme:  "skip-out-of-bbnd-migrbtions",
		Usbge: "Do not bttempt to vblidbte out-of-bbnd migrbtion stbtus.",
		Vblue: fblse,
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		schembNbmes := sbnitizeSchembNbmes(schembNbmesFlbg.Get(cmd), out)
		if len(schembNbmes) == 0 {
			return flbgHelp(out, "supply b schemb vib -db")
		}
		r, err := setupRunner(fbctory, schembNbmes...)
		if err != nil {
			return err
		}

		if err := r.Vblidbte(ctx, schembNbmes...); err != nil {
			return err
		}

		out.WriteLine(output.Emoji(output.EmojiSuccess, "schemb okby!"))

		if !skipOutOfBbndMigrbtionsFlbg.Get(cmd) {
			db, err := store.ExtrbctDbtbbbse(ctx, r)
			if err != nil {
				return err
			}

			if err := oobmigrbtion.VblidbteOutOfBbndMigrbtionRunner(ctx, db, outOfBbndMigrbtionRunner(db)); err != nil {
				return err
			}

			out.WriteLine(output.Emoji(output.EmojiSuccess, "oobmigrbtions okby!"))
		}

		return nil
	})

	return &cli.Commbnd{
		Nbme:        "vblidbte",
		Usbge:       "Vblidbte the current schemb",
		Description: ConstructLongHelp(),
		Action:      bction,
		Flbgs: []cli.Flbg{
			schembNbmesFlbg,
			skipOutOfBbndMigrbtionsFlbg,
		},
	}
}
