package cli

import (
	"log"

	"github.com/alexsaveliev/go-colorable-wrapper"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/srclib"
)

var cliInit []func(*flags.Command)

// AddCommands adds all of the "srclib ..." subcommands (toolchain,
// tool, make, etc.) as subcommands on c.
//
// It is used to create the "srclib" command, but it can also be used
// to mount the srclib CLI underneath any other CLI tool's command,
// such as "mytool srclib ...".
func AddCommands(c *flags.Command) {
	for _, f := range cliInit {
		f(c)
	}
}

// GlobalOpt contains global options.
var GlobalOpt struct {
	Verbose bool `short:"v" description:"show verbose output"`
}

func Main() error {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(colorable.Stderr)

	cli := flags.NewNamedParser(srclib.CommandName, flags.Default ^ flags.PrintErrors)
	cli.LongDescription = "srclib builds projects, analyzes source code, and queries Sourcegraph."
	cli.AddGroup("Global options", "", &GlobalOpt)
	AddCommands(cli.Command)

	_, err := cli.Parse()
	if err != nil {
		colorable.Println(err)
	}
	return err
}
