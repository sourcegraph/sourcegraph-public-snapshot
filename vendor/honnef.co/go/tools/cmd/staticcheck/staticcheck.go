// staticcheck detects a myriad of bugs and inefficiencies in your
// code.
package main // import "honnef.co/go/tools/cmd/staticcheck"

import (
	"os"

	"honnef.co/go/tools/lint/lintutil"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	fs := lintutil.FlagSet("staticcheck")
	gen := fs.Bool("generated", false, "Check generated code")
	fs.Parse(os.Args[1:])
	c := staticcheck.NewChecker()
	c.CheckGenerated = *gen
	cfg := lintutil.CheckerConfig{
		Checker:     c,
		ExitNonZero: true,
	}
	lintutil.ProcessFlagSet([]lintutil.CheckerConfig{cfg}, fs)
}
