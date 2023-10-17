package workspace

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/files"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
)

type dockerWorkspace struct {
	scriptFilenames []string
	workspaceDir    string
	logger          cmdlogger.Logger
}

// NewDockerWorkspace creates a new workspace for docker-based execution. A path on
// the host will be used to set up the workspace, clone the repo and put script files.
func NewDockerWorkspace(
	ctx context.Context,
	filesStore files.Store,
	job types.Job,
	cmd command.Command,
	logger cmdlogger.Logger,
	cloneOpts CloneOptions,
	operations *command.Operations,
) (Workspace, error) {
	workspaceDir, err := makeTemporaryDirectory("workspace-" + strconv.Itoa(job.ID))
	if err != nil {
		return nil, err
	}

	if job.RepositoryName != "" {
		if err = cloneRepo(ctx, workspaceDir, job, cmd, logger, cloneOpts, operations); err != nil {
			_ = os.RemoveAll(workspaceDir)
			return nil, err
		}
	}

	scriptPaths, err := prepareScripts(ctx, filesStore, job, workspaceDir, logger)
	if err != nil {
		_ = os.RemoveAll(workspaceDir)
		return nil, err
	}

	return &dockerWorkspace{
		scriptFilenames: scriptPaths,
		workspaceDir:    workspaceDir,
		logger:          logger,
	}, nil
}

func makeTemporaryDirectory(prefix string) (string, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return "", err
		}
		return os.MkdirTemp(tempdir, prefix+"-*")
	}

	return os.MkdirTemp("", prefix+"-*")
}

func (w dockerWorkspace) Path() string {
	return w.workspaceDir
}

func (w dockerWorkspace) WorkingDirectory() string {
	return w.workspaceDir
}

func (w dockerWorkspace) ScriptFilenames() []string {
	return w.scriptFilenames
}

func (w dockerWorkspace) Remove(ctx context.Context, keepWorkspace bool) {
	handle := w.logger.LogEntry("teardown.fs", nil)
	defer func() {
		// We always finish this with exit code 0 even if it errored, because workspace
		// cleanup doesn't fail the execution job. We can deal with it separately.
		handle.Finalize(0)
		handle.Close()
	}()

	if keepWorkspace {
		fmt.Fprintf(handle, "Preserving workspace (%s) as per config", w.workspaceDir)
		return
	}

	fmt.Fprintf(handle, "Removing %s\n", w.workspaceDir)
	if rmErr := os.RemoveAll(w.workspaceDir); rmErr != nil {
		fmt.Fprintf(handle, "Operation failed: %s\n", rmErr.Error())
	}
}
