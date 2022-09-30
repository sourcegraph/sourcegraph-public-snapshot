package run

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RunTestVM(cliCtx *cli.Context, logger log.Logger, config *config.Config) error {
	repoName := cliCtx.String("repo")
	revision := cliCtx.String("revision")
	nameOnly := cliCtx.Bool("name-only")

	if repoName != "" && revision == "" {
		return errors.New("must specify revision when setting --repo")
	}

	var logOutput io.Writer = os.Stdout
	if nameOnly {
		logOutput = os.Stderr
	}
	name, err := createVM(cliCtx.Context, config, repoName, revision, logOutput)
	if err != nil {
		return err
	}

	if nameOnly {
		fmt.Print(name)
	} else {
		fmt.Printf("Success! Connect to the VM using\n  $ ignite attach %s\n\nOnce done run\n  $ ignite rm --force %s\nto clean up the running VM.\n", name, name)
	}

	return nil
}

func createVM(ctx context.Context, config *config.Config, repositoryName, revision string, logOutput io.Writer) (string, error) {
	vmNameSuffix, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	// Use a static prefix, so these VMs aren't cleaned up by a running executor
	// VM janitor.
	name := fmt.Sprintf("%s-%s", "executor-test-vm", vmNameSuffix.String())

	commandLogger := command.NewWriterLogger(logOutput)
	operations := command.NewOperations(&observation.TestContext)

	hostRunner := command.NewRunner("", commandLogger, command.Options{}, operations)
	workspace, err := workspace.NewFirecrackerWorkspace(
		ctx,
		// Just enough to spin up a VM.
		executor.Job{
			RepositoryName: repositoryName,
			Commit:         revision,
		},
		config.FirecrackerDiskSpace,
		// Always keep the workspace in this debug command.
		true,
		hostRunner,
		commandLogger,
		// TODO: get git service path from config.
		workspace.CloneOptions{
			EndpointURL:    config.FrontendURL,
			GitServicePath: "/.executors/git",
			ExecutorToken:  config.FrontendAuthorizationToken,
		},
		operations,
	)
	if err != nil {
		return "", err
	}

	fopts := firecrackerOptions(config)
	fopts.Enabled = true

	runner := command.NewRunner(workspace.Path(), commandLogger, command.Options{
		ExecutorName:       name,
		ResourceOptions:    resourceOptions(config),
		FirecrackerOptions: fopts,
	}, operations)

	if err := runner.Setup(ctx); err != nil {
		return "", err
	}

	return name, nil
}
