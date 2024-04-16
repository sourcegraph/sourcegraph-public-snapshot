package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/grafana/regexp"

	"github.com/mattn/go-isatty"
)

// Returns the string for a foreground ANSI 8 bit color code.
func fg256Color(code int) string {
	return fmt.Sprintf("\033[38;5;%dm", code)
}

// Returns the string for a background ANSI 8 bit color code.
func bg256Color(code int) string {
	return fmt.Sprintf("\033[48;5;%dm", code)
}

// See https://i.stack.imgur.com/KTSQa.png or https://jonasjacek.github.io/colors/
var ansiColors = map[string]string{
	"nc":      "\033[0m",
	"logo":    fg256Color(57),
	"warning": fg256Color(124),
	"success": fg256Color(2),

	// Search-specific colors.
	"search-query":          fg256Color(68),
	"search-border":         fg256Color(239),
	"search-link":           fg256Color(237),
	"search-repository":     fg256Color(23),
	"search-branch":         fg256Color(0) + bg256Color(7),
	"search-filename":       fg256Color(69),
	"search-match":          fg256Color(0) + bg256Color(11),
	"search-line-numbers":   fg256Color(69),
	"search-commit-author":  fg256Color(2),
	"search-commit-subject": fg256Color(68),
	"search-commit-date":    fg256Color(23),

	// Search alert specific colors.
	"search-alert-title":                fg256Color(124),
	"search-alert-description":          fg256Color(124),
	"search-alert-proposed-title":       "",
	"search-alert-proposed-query":       fg256Color(69),
	"search-alert-proposed-description": "",
}

// Borrowed from https://github.com/acarl005/stripansi/blob/master/stripansi.go
// MIT licensed, see https://github.com/acarl005/stripansi/blob/master/LICENSE
var ansiRegexp = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")

var isTest bool
var colorDisabled bool

func init() {
	if !isTest {
		// We comply with the no-color.org spec here.
		if os.Getenv("NO_COLOR") != "" {
			colorDisabled = true
		} else {
			// If they specify COLOR=true or COLOR=false, we respect that.
			if color := os.Getenv("COLOR"); color != "" {
				colorEnabled, _ := strconv.ParseBool(color)
				colorDisabled = !colorEnabled
			} else {
				// If our program is being piped into another one, then disable
				// color. This is usually desired, and can be overridden with COLOR=true.
				colorDisabled = !isatty.IsTerminal(os.Stdout.Fd())
			}
		}
	}
	if colorDisabled {
		for color := range ansiColors {
			ansiColors[color] = ""
		}
	}

	if os.Getenv("DEBUG_PRINT_COLORS") == "t" {
		fmt.Println("The following colors are available:")
		for color, code := range ansiColors {
			if color == "nc" {
				continue
			}
			fmt.Println(code + color + ansiColors["nc"])
		}
		os.Exit(1)
	}
}
