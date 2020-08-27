package output

import (
	"os"
	"strconv"

	"github.com/mattn/go-isatty"
	"github.com/nsf/termbox-go"
)

type capabilities struct {
	Color  bool
	Isatty bool
	Height int
	Width  int
}

func detectCapabilities() capabilities {
	// There's a pretty obvious flaw here in that we only check the terminal
	// size once. We may want to consider adding a background goroutine that
	// updates the capabilities struct every second or two.
	//
	// Pulling in termbox is probably overkill, but finding a pure Go library
	// that could just provide terminfo was surprisingly hard. At least termbox
	// is widely used.
	w, h := 80, 25
	if err := termbox.Init(); err == nil {
		w, h = termbox.Size()
		termbox.Close()
	}

	atty := isatty.IsTerminal(os.Stdout.Fd())

	return capabilities{
		Color:  detectColor(atty),
		Isatty: atty,
		Height: h,
		Width:  w,
	}
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
