package cli

import (
	"log"
	"os"

	"sourcegraph.com/sourcegraph/makex"
)

func init() {
	_, err := CLI.AddCommand("makefile",
		"prints the Makefile that the `make` subcommand executes",
		"The makefile command prints the Makefile that the `make` subcommand will execute.",
		&makefileCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

type MakefileCmd struct {
	ToolchainExecOpt `group:"execution"`
	Verbose          bool `short:"v" long:"verbose" description:"show more verbose output"`
}

var makefileCmd MakefileCmd

func (c *MakefileCmd) Execute(args []string) error {
	mf, err := CreateMakefile(c.ToolchainExecOpt, c.Verbose)
	if err != nil {
		return err
	}

	mfData, err := makex.Marshal(mf)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(mfData)
	return err
}
