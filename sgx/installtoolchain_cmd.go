package sgx

import (
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	// Add an internal "srclib toolchain" subcommand shortcut.
	cli.CLI.AddExistingCommand(srclib.CLI.Find("toolchain"))
}
