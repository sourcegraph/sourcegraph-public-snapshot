package main // import "honnef.co/go/tools/cmd/errcheck-ng"

import (
	"os"

	"honnef.co/go/tools/errcheck"
	"honnef.co/go/tools/lint/lintutil"
)

func main() {
	c := errcheck.NewChecker()
	lintutil.ProcessArgs("errcheck-ng", c, os.Args[1:])
}
