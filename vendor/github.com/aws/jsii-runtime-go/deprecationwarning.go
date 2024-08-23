//go:build (go1.16 || go1.17) && !go1.18
// +build go1.16 go1.17
// +build !go1.18

package jsii

import (
	"fmt"
	"os"

	"github.com/aws/jsii-runtime-go/internal/compiler"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// / Emits a deprecation warning message when
func init() {
	tty := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())

	if tty {
		// Set terminal to bold red
		color.Set(color.FgRed, color.Bold)
	}

	fmt.Fprintln(os.Stderr, "###########################################################")
	fmt.Fprintf(os.Stderr, "# This binary was compiled with %v, which has reached #\n", compiler.Version)
	fmt.Fprintf(os.Stderr, "# end-of-life on %v.                              #\n", compiler.EndOfLifeDate)
	fmt.Fprintln(os.Stderr, "#                                                         #")
	fmt.Fprintln(os.Stderr, "# Support for this version WILL be dropped in the future! #")
	fmt.Fprintln(os.Stderr, "#                                                         #")
	fmt.Fprintln(os.Stderr, "# See https://go.dev/security for more information.       #")
	fmt.Fprintln(os.Stderr, "###########################################################")

	if tty {
		// Reset terminal back to normal
		color.Unset()
	}
}
