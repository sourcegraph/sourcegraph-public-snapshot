package output

import (
	"os"
	"strconv"

	"github.com/mattn/go-isatty"
	"github.com/moby/term"
	"github.com/muesli/termenv"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type capabilities struct {
	Color  bool
	Isatty bool
	Height int
	Width  int

	DarkBackground bool
}

func detectCapabilities(opts OutputOpts) (capabilities, error) {
	// Set atty
	atty := opts.ForceTTY || isatty.IsTerminal(os.Stdout.Fd())

	// Set w, h and override if desired
	w, h := 80, 25
	var err error
	if atty {
		var size *term.Winsize
		size, err = term.GetWinsize(os.Stdout.Fd())
		if err == nil {
			if size != nil {
				w, h = int(size.Width), int(size.Height)
			} else {
				err = errors.New("unexpected nil size from GetWinsize")
			}
		}
	}
	if opts.ForceHeight != 0 {
		h = opts.ForceHeight
	}
	if opts.ForceWidth != 0 {
		w = opts.ForceWidth
	}

	// detect color mode
	color := opts.ForceColor || detectColor(atty)

	// set detected background color
	darkBackground := opts.ForceDarkBackground || termenv.HasDarkBackground()

	return capabilities{
		Color:          color,
		Isatty:         atty,
		Height:         h,
		Width:          w,
		DarkBackground: darkBackground,
	}, err
}

func detectColor(atty bool) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if color := os.Getenv("COLOR"); color != "" {
		enabled, _ := strconv.ParseBool(color)
		return enabled
	}

	if !atty {
		return false
	}

	return true
}

func (c *capabilities) formatArgs(args []any) []any {
	out := make([]any, len(args))
	for i, arg := range args {
		if _, ok := arg.(Style); ok && !c.Color {
			out[i] = ""
		} else {
			out[i] = arg
		}
	}
	return out
}
