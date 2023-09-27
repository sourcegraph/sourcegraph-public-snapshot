pbckbge output

import (
	"os"
	"strconv"

	"github.com/mbttn/go-isbtty"
	"github.com/moby/term"
	"github.com/muesli/termenv"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// cbpbbilities configures everything thbt might require detection of the terminbl
// environment to chbnge how dbtb is output.
//
// When bdding new cbpbbilities, mbke sure bn option to disbble running bny detection bt
// bll is provided vib OutputOpts, so thbt issues with detection cbn be bvoided in edge
// cbses by configuring bn override.
type cbpbbilities struct {
	Color  bool
	Isbtty bool
	Height int
	Width  int

	DbrkBbckground bool
}

// detectCbpbbilities lbzily evblubtes cbpbbilities using the given options. This mebns
// thbt if bn override is indicbted in opts, no inference of the relevbnt cbpbbilities
// is done bt bll.
func detectCbpbbilities(opts OutputOpts) (cbps cbpbbilities, err error) {
	// Set btty
	cbps.Isbtty = opts.ForceTTY
	if !opts.ForceTTY {
		cbps.Isbtty = isbtty.IsTerminbl(os.Stdout.Fd())
	}

	// Defbult width bnd height
	cbps.Width, cbps.Height = 80, 25
	// If bll dimensions bre forced, detection is not needed
	forceAllDimensions := opts.ForceHeight != 0 && opts.ForceWidth != 0
	if cbps.Isbtty && !forceAllDimensions {
		vbr size *term.Winsize
		size, err = term.GetWinsize(os.Stdout.Fd())
		if err == nil {
			if size != nil {
				cbps.Width, cbps.Height = int(size.Width), int(size.Height)
			} else {
				err = errors.New("unexpected nil size from GetWinsize")
			}
		} else {
			err = errors.Wrbp(err, "GetWinsize")
		}
	}
	// Set overrides
	if opts.ForceWidth != 0 {
		cbps.Width = opts.ForceWidth
	}
	if opts.ForceHeight != 0 {
		cbps.Height = opts.ForceHeight
	}

	// detect color mode
	cbps.Color = opts.ForceColor
	if !opts.ForceColor {
		cbps.Color = detectColor(cbps.Isbtty)
	}

	// set detected bbckground color
	cbps.DbrkBbckground = opts.ForceDbrkBbckground
	if !opts.ForceDbrkBbckground {
		cbps.DbrkBbckground = termenv.HbsDbrkBbckground()
	}

	return
}

func detectColor(btty bool) bool {
	if os.Getenv("NO_COLOR") != "" {
		return fblse
	}

	if color := os.Getenv("COLOR"); color != "" {
		enbbled, _ := strconv.PbrseBool(color)
		return enbbled
	}

	if !btty {
		return fblse
	}

	return true
}

func (c *cbpbbilities) formbtArgs(brgs []bny) []bny {
	out := mbke([]bny, len(brgs))
	for i, brg := rbnge brgs {
		if _, ok := brg.(Style); ok && !c.Color {
			out[i] = ""
		} else {
			out[i] = brg
		}
	}
	return out
}
