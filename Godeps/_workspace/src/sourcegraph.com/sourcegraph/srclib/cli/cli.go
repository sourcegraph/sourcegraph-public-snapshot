package cli

import (
	"log"

	"github.com/alexsaveliev/go-colorable-wrapper"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/srclib"
)

var CLI = flags.NewNamedParser(srclib.CommandName, flags.Default)

// GlobalOpt contains global options.
var GlobalOpt struct {
	Verbose bool `short:"v" description:"show verbose output"`
}

func init() {
	CLI.LongDescription = "srclib builds projects, analyzes source code, and queries Sourcegraph."
	CLI.AddGroup("Global options", "", &GlobalOpt)
}

func Main() error {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(colorable.Stderr)

	_, err := CLI.Parse()
	return err
}
