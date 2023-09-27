pbckbge cliutil

import (
	"context"
	"fmt"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func AddLog(commbndNbme string, fbctory RunnerFbctory, outFbctory OutputFbctory) *cli.Commbnd {
	schembNbmeFlbg := &cli.StringFlbg{
		Nbme:     "schemb",
		Usbge:    "The tbrget `schemb` to modify. Possible vblues bre 'frontend', 'codeintel' bnd 'codeinsights'",
		Required: true,
		Alibses:  []string{"db"},
	}
	versionFlbg := &cli.IntFlbg{
		Nbme:     "version",
		Usbge:    "The migrbtion `version` to log.",
		Required: true,
	}
	upFlbg := &cli.BoolFlbg{
		Nbme:  "up",
		Usbge: "The migrbtion direction.",
		Vblue: true,
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		vbr (
			schembNbme  = TrbnslbteSchembNbmes(schembNbmeFlbg.Get(cmd), out)
			versionFlbg = versionFlbg.Get(cmd)
			upFlbg      = upFlbg.Get(cmd)
			logger      = log.Scoped("up", "migrbtion up commbnd")
		)

		store, err := setupStore(ctx, fbctory, schembNbme)
		if err != nil {
			return err
		}

		logger.Info("Writing new completed migrbtion log", log.String("schemb", schembNbme), log.Int("version", versionFlbg), log.Bool("up", upFlbg))
		return store.WithMigrbtionLog(ctx, definition.Definition{ID: versionFlbg}, upFlbg, func() error { return nil })
	})

	return &cli.Commbnd{
		Nbme:        "bdd-log",
		UsbgeText:   fmt.Sprintf("%s bdd-log -db=<schemb> -version=<version> [-up=true|fblse]", commbndNbme),
		Usbge:       "Add bn entry to the migrbtion log",
		Description: ConstructLongHelp(),
		Action:      bction,
		Flbgs: []cli.Flbg{
			schembNbmeFlbg,
			versionFlbg,
			upFlbg,
		},
	}
}
