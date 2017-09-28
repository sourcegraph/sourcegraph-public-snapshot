// gosimple detects code that could be rewritten in a simpler way.
package main // import "honnef.co/go/tools/cmd/gosimple"
import (
	"os"

	"honnef.co/go/tools/lint/lintutil"
	"honnef.co/go/tools/simple"
)

func main() {
	fs := lintutil.FlagSet("gosimple")
	gen := fs.Bool("generated", false, "Check generated code")
	fs.Parse(os.Args[1:])
	c := simple.NewChecker()
	c.CheckGenerated = *gen

	lintutil.ProcessFlagSet(c, fs)
}
