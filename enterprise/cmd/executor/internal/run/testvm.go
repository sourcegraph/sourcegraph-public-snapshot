package run

import (
	"fmt"

	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func RunTestVM(cliCtx *cli.Context, logger log.Logger, config *config.Config) error {
	repoName := cliCtx.String("repo")
	nameOnly := cliCtx.Bool("name-only")

	commandLogger := command.NewNoopLogger()
	operations := command.NewOperations(&observation.TestContext)

	hostRunner := command.NewRunner("", commandLogger, command.Options{}, operations)
	workspace, err := workspace.NewFirecrackerWorkspace(
		cliCtx.Context,
		// Just enough to spin up a VM.
		executor.Job{
			ID:                  123,
			RepositoryName:      repoName,
			RepositoryDirectory: "repo",
			Commit:              "HEAD",
		},
		config.FirecrackerDiskSpace,
		// Always keep the workspace in this debug command.
		true,
		hostRunner,
		commandLogger,
		workspace.CloneOptions{
			EndpointURL: config.FrontendURL,
			// TODO: Validate this is correct.
			GitServicePath: ".api/executor/",
			ExecutorToken:  config.FrontendAuthorizationToken,
		},
		operations,
	)
	if err != nil {
		return err
	}

	runner := command.NewRunner(workspace.Path(), commandLogger, command.Options{
		// TODO: Use helper to create firecracker options object.
		FirecrackerOptions: command.FirecrackerOptions{
			Enabled:                 true,
			Image:                   config.FirecrackerImage,
			KernelImage:             config.FirecrackerKernelImage,
			SandboxImage:            config.FirecrackerSandboxImage,
			VMStartupScriptPath:     config.VMStartupScriptPath,
			DockerRegistryMirrorURL: config.DockerRegistryMirrorURL,
		},
	}, command.NewOperations(&observation.TestContext))

	fmt.Printf("Spawning ignite VM with %s cloned into the workspace...\n", repoName)

	if err := runner.Setup(cliCtx.Context); err != nil {
		return err
	}

	// TODO: Get the VM name from the runner.
	if nameOnly {
		fmt.Printf("executor-debug-vm-deadbeef")
	} else {
		fmt.Printf("Success! Connect to the VM using\n  $ ignite attach executor-debug-vm-deadbeef\n\nOnce done run\n  $ ignite rm --force executor-debug-vm-deadbeef\nto clean up the running VM.\n")
	}

	return nil
}
