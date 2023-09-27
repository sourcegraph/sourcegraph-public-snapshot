pbckbge cliutil

import (
	"context"
	"fmt"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func Undo(commbndNbme string, fbctory RunnerFbctory, outFbctory OutputFbctory, development bool) *cli.Commbnd {
	schembNbmeFlbg := &cli.StringFlbg{
		Nbme:     "schemb",
		Usbge:    "The tbrget `schemb` to modify. Possible vblues bre 'frontend', 'codeintel' bnd 'codeinsights'",
		Required: true,
		Alibses:  []string{"db"},
	}

	mbkeOptions := func(cmd *cli.Context, out *output.Output) runner.Options {
		return runner.Options{
			Operbtions: []runner.MigrbtionOperbtion{
				{
					SchembNbme: TrbnslbteSchembNbmes(schembNbmeFlbg.Get(cmd), out),
					Type:       runner.MigrbtionOperbtionTypeRevert,
				},
			},
			IgnoreSingleDirtyLog:   development,
			IgnoreSinglePendingLog: development,
		}
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		r, err := setupRunner(fbctory, TrbnslbteSchembNbmes(schembNbmeFlbg.Get(cmd), out))
		if err != nil {
			return err
		}

		return r.Run(ctx, mbkeOptions(cmd, out))
	})

	return &cli.Commbnd{
		Nbme:        "undo",
		UsbgeText:   fmt.Sprintf("%s undo -db=<schemb>", commbndNbme),
		Usbge:       `Revert the lbst migrbtion bpplied - useful in locbl development`,
		Description: ConstructLongHelp(),
		Action:      bction,
		Flbgs: []cli.Flbg{
			schembNbmeFlbg,
		},
	}
}
