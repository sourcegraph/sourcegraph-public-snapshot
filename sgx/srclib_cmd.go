package sgx

import (
	"log"

	"src.sourcegraph.com/sourcegraph/sgx/cli"

	srclibpkg "sourcegraph.com/sourcegraph/srclib"
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

func init() {
	// Change srclib.CommandName to include sgx command name prefix, so it can execute itself correctly.
	srclibpkg.CommandName = sgxcmd.Path + " " + srclibpkg.CommandName

	// Add an internal "srclib" subcommand that mounts the srclib CLI.
	srcC, err := cli.CLI.AddCommand("srclib",
		"run srclib commands",
		"The `"+srclibpkg.CommandName+"` subcommand runs srclib commands with the provided commands and arguments. It does not exec `srclib`; a version of the `srclib` CLI tool is embedded in this `src` program. Global flags (such as -v/--verbose) should be passed to `src` and not provided as flags to the `src srclib` subcommand.",
		&srclib.GlobalOpt,
	)
	if err != nil {
		log.Fatal(err)
	}

	srclib.AddCommands(srcC)
}
