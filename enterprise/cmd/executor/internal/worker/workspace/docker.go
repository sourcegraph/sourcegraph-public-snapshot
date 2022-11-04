package workspace

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
)

// NewDockerWorkspace creates a new workspace for docker-based execution. A path on
// the host will be used to set up the workspace, clone the repo and put script files.
func NewDockerWorkspace(
	ctx context.Context,
	filesStore store.FilesStore,
	job executor.Job,
	commandRunner command.Runner,
	logger command.Logger,
	cloneOpts CloneOptions,
	operations *command.Operations,
) (Workspace, error) {
	workspaceDir, err := MakeTempDirectory("workspace-" + strconv.Itoa(job.ID))
	if err != nil {
		return nil, err
	}

	if job.RepositoryName != "" {
		if err := cloneRepo(ctx, workspaceDir, job, commandRunner, cloneOpts, operations); err != nil {
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
		path:            workspaceDir,
		scriptFilenames: scriptPaths,
		workspaceDir:    workspaceDir,
		logger:          logger,
	}, nil
}

type dockerWorkspace struct {
	path            string
	scriptFilenames []string
	workspaceDir    string
	logger          command.Logger
}

func (w dockerWorkspace) Path() string {
	return w.path
}

func (w dockerWorkspace) ScriptFilenames() []string {
	return w.scriptFilenames
}

func (w dockerWorkspace) Remove(ctx context.Context, keepWorkspace bool) {
	handle := w.logger.Log("teardown.fs", nil)
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
