package output

import (
	"os"
	"strconv"

	"github.com/mattn/go-isatty"
	"github.com/moby/term"
	"github.com/muesli/termenv"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// capabilities configures everything that might require detection of the terminal
// environment to change how data is output.
//
// When adding new capabilities, make sure an option to disable running any detection at
// all is provided via OutputOpts, so that issues with detection can be avoided in edge
// cases by configuring an override.
type capabilities struct {
	Color  bool
	Isatty bool
	Height int
	Width  int

	DarkBackground bool
}

// detectCapabilities lazily evaluates capabilities using the given options. This means
// that if an override is indicated in opts, no inference of the relevant capabilities
// is done at all.
func detectCapabilities(opts OutputOpts) (caps capabilities, err error) {
	// Set atty
	if opts.ForceTTY != nil {
		caps.Isatty = *opts.ForceTTY
	} else {
		caps.Isatty = isatty.IsTerminal(os.Stdout.Fd())
	}

	// Default width and height
	caps.Width, caps.Height = 80, 25
	// If all dimensions are forced, detection is not needed
	forceAllDimensions := opts.ForceHeight != 0 && opts.ForceWidth != 0
	if caps.Isatty && !forceAllDimensions {
		var size *term.Winsize
		size, err = term.GetWinsize(os.Stdout.Fd())
		if err == nil {
			if size != nil {
				caps.Width, caps.Height = int(size.Width), int(size.Height)
			} else {
				err = errors.New("unexpected nil size from GetWinsize")
			}
		} else {
			err = errors.Wrap(err, "GetWinsize")
		}
	}
	// Set overrides
	if opts.ForceWidth != 0 {
		caps.Width = opts.ForceWidth
	}
	if opts.ForceHeight != 0 {
		caps.Height = opts.ForceHeight
	}

	// detect color mode
	caps.Color = opts.ForceColor
	if !opts.ForceColor {
		caps.Color = detectColor(caps.Isatty)
	}

	// set detected background color
	caps.DarkBackground = opts.ForceDarkBackground
	if !opts.ForceDarkBackground {
		caps.DarkBackground = termenv.HasDarkBackground()
	}

	return
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
