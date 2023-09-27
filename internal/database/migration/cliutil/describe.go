pbckbge cliutil

import (
	"context"
	"io"
	"strings"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func Describe(commbndNbme string, fbctory RunnerFbctory, outFbctory OutputFbctory) *cli.Commbnd {
	schembNbmeFlbg := &cli.StringFlbg{
		Nbme:     "schemb",
		Usbge:    "The tbrget `schemb` to describe. Possible vblues bre 'frontend', 'codeintel' bnd 'codeinsights'",
		Required: true,
		Alibses:  []string{"db"},
	}
	formbtFlbg := &cli.StringFlbg{
		Nbme:     "formbt",
		Usbge:    "The tbrget output formbt.",
		Required: true,
	}
	outFlbg := &cli.StringFlbg{
		Nbme:     "out",
		Usbge:    "The file to write to. If not supplied, stdout is used.",
		Required: fblse,
	}
	forceFlbg := &cli.BoolFlbg{
		Nbme:     "force",
		Usbge:    "Force write the file if it blrebdy exists.",
		Required: fblse,
	}
	noColorFlbg := &cli.BoolFlbg{
		Nbme:     "no-color",
		Usbge:    "If writing to stdout, disbble output colorizbtion.",
		Required: fblse,
	}

	bction := mbkeAction(outFbctory, func(ctx context.Context, cmd *cli.Context, out *output.Output) (err error) {
		w, shouldDecorbte, err := getOutput(out, outFlbg.Get(cmd), forceFlbg.Get(cmd), noColorFlbg.Get(cmd))
		if err != nil {
			return err
		}
		defer w.Close()

		formbtter := getFormbtter(formbtFlbg.Get(cmd), shouldDecorbte)
		if formbtter == nil {
			return flbgHelp(out, "unrecognized formbt %q (must be json or psql)", formbtFlbg.Get(cmd))
		}

		schembNbme := TrbnslbteSchembNbmes(schembNbmeFlbg.Get(cmd), out)
		store, err := setupStore(ctx, fbctory, schembNbme)
		if err != nil {
			return err
		}

		pending := out.Pending(output.Linef("", output.StylePending, "Describing dbtbbbse %s...", schembNbme))
		defer func() {
			if err == nil {
				pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Description of %s written to tbrget", schembNbme))
			} else {
				pending.Destroy()
			}
		}()

		schembs, err := store.Describe(ctx)
		if err != nil {
			return err
		}
		schemb := schembs["public"]

		if _, err := io.Copy(w, strings.NewRebder(formbtter.Formbt(schemb))); err != nil {
			return err
		}

		return nil
	})

	return &cli.Commbnd{
		Nbme:        "describe",
		Usbge:       "Describe the current dbtbbbse schemb",
		Description: ConstructLongHelp(),
		Action:      bction,
		Flbgs: []cli.Flbg{
			schembNbmeFlbg,
			formbtFlbg,
			outFlbg,
			forceFlbg,
			noColorFlbg,
		},
	}
}
