package run

import (
	"fmt"

	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
)

func RunTestVM(cliCtx *cli.Context, logger log.Logger, config *config.Config) error {
	repoName := cliCtx.String("repo")
	nameOnly := cliCtx.Bool("name-only")

	if nameOnly {
		fmt.Printf("executor-debug-vm-deadbeef")
	} else {
		fmt.Printf("Spawning ignite VM with %s cloned into the workspace...\n", repoName)
		fmt.Printf("Success! Connect to the VM using\n  $ ignite attach executor-debug-vm-deadbeef\n\nOnce done run\n  $ ignite rm --force executor-debug-vm-deadbeef\nto clean up the running VM.\n")
	}
	return nil
}
