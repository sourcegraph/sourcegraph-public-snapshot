package run

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/workspace"
	internalexecutor "github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TestVM is the CLI action handler for the test-vm command. It spawns a firecracker
// VM for testing purposes.
//
// TODO: Add a command to get rid of VM without calling ignite, this way we can inline or replace ignite later
// more easily.
// TODO: Add a command to attach to the VM without calling ignite, this way we can inline or replace ignite later
// more easily.
func TestVM(cliCtx *cli.Context, cmdRunner util.CmdRunner, logger log.Logger, config *config.Config) error {
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
	name, err := createVM(cliCtx.Context, cmdRunner, config, repoName, revision, logOutput)
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

func createVM(ctx context.Context, cmdRunner util.CmdRunner, config *config.Config, repositoryName, revision string, logOutput io.Writer) (string, error) {
	vmNameSuffix, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	// Use a static prefix, so these VMs aren't cleaned up by a running executor
	// VM janitor.
	name := fmt.Sprintf("%s-%s", "executor-test-vm", vmNameSuffix.String())

	commandLogger := &writerLogger{w: logOutput}
	operations := command.NewOperations(&observation.TestContext)

	cmd := &command.RealCommand{
		CmdRunner: cmdRunner,
		Logger:    log.Scoped("executor-test-vm"),
	}
	firecrackerWorkspace, err := workspace.NewFirecrackerWorkspace(
		ctx,
		// No need for files store in the test.
		nil,
		// Just enough to spin up a VM.
		types.Job{
			RepositoryName: repositoryName,
			Commit:         revision,
		},
		config.FirecrackerDiskSpace,
		// Always keep the workspace in this debug command.
		true,
		cmdRunner,
		cmd,
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

	firecrackerRunner := runner.NewFirecrackerRunner(cmd, commandLogger, firecrackerWorkspace.Path(), name, fopts, types.DockerAuthConfig{}, operations)

	if err = firecrackerRunner.Setup(ctx); err != nil {
		return "", err
	}

	return name, nil
}

type writerLogger struct {
	w io.Writer
}

func (*writerLogger) Flush() error { return nil }

func (l *writerLogger) LogEntry(key string, command []string) cmdlogger.LogEntry {
	fmt.Fprintf(l.w, "%s: %s", key, strings.Join(command, " "))
	return &writerLogEntry{w: l.w}
}

type writerLogEntry struct {
	w io.Writer
}

func (l *writerLogEntry) Write(p []byte) (n int, err error) {
	return fmt.Fprint(l.w, string(p))
}

func (*writerLogEntry) Close() error { return nil }

func (*writerLogEntry) Finalize(exitCode int) {}

func (*writerLogEntry) CurrentLogEntry() internalexecutor.ExecutionLogEntry {
	return internalexecutor.ExecutionLogEntry{}
}
