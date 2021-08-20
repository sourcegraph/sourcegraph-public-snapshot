package output

import (
	"os"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/mattn/go-isatty"
	"github.com/moby/term"
)

type capabilities struct {
	Color  bool
	Isatty bool
	Height int
	Width  int
}

func detectCapabilities() (capabilities, error) {
	atty := isatty.IsTerminal(os.Stdout.Fd())
	w, h := 80, 25
	var err error
	if atty {
		size, err := term.GetWinsize(os.Stdout.Fd())
		if err == nil {
			if size != nil {
				w, h = int(size.Width), int(size.Height)
			} else {
				err = errors.New("unexpected nil size from GetWinsize")
			}
		}
	}

	return capabilities{
		Color:  detectColor(atty),
		Isatty: atty,
		Height: h,
		Width:  w,
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

func (c *capabilities) formatArgs(args []interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, arg := range args {
		if _, ok := arg.(Style); ok && !c.Color {
			out[i] = ""
		} else {
			out[i] = arg
		}
	}
	return out
}
