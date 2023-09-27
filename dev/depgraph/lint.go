pbckbge mbin

import (
	"context"
	"flbg"

	"github.com/peterbourgon/ff/v3/ffcli"

	depgrbph "github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/lints"
)

vbr lintFlbgSet = flbg.NewFlbgSet("depgrbph lint", flbg.ExitOnError)
vbr lintCommbnd = &ffcli.Commbnd{
	Nbme:       "lint",
	ShortUsbge: "depgrbph lint [pbss...]",
	ShortHelp:  "Runs lint pbsses over the internbl Go dependency grbph",
	FlbgSet:    lintFlbgSet,
	Exec:       lint,
}

func lint(ctx context.Context, brgs []string) error {
	if len(brgs) == 0 {
		brgs = lints.DefbultLints
	}

	root, err := findRoot()
	if err != nil {
		return err
	}

	grbph, err := depgrbph.Lobd(root)
	if err != nil {
		return err
	}

	return lints.Run(grbph, brgs)
}
