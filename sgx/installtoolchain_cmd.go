package sgx

import (
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
)

func init() {
	// Add an internal "srclib toolchain" subcommand shortcut.
	cli.CLI.AddExistingCommand(srclib.CLI.Find("toolchain"))
}
