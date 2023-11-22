package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/files"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
)

type kubernetesWorkspace struct {
	scriptFilenames []string
	workspaceDir    string
	logger          cmdlogger.Logger
}

// NewKubernetesWorkspace creates a new workspace for a job.
func NewKubernetesWorkspace(
	ctx context.Context,
	filesStore files.Store,
	job types.Job,
	cmd command.Command,
	logger cmdlogger.Logger,
	cloneOpts CloneOptions,
	mountPath string,
	singleJob bool,
	operations *command.Operations,
) (Workspace, error) {
	// TODO switch to the single job in 5.2
	if singleJob {
		return &kubernetesWorkspace{logger: logger}, nil
	}

	workspaceDir := filepath.Join(mountPath, fmt.Sprintf("job-%d", job.ID))

	if err := os.MkdirAll(workspaceDir, os.ModePerm); err != nil {
		return nil, err
	}

	if job.RepositoryName != "" {
		if err := cloneRepo(ctx, workspaceDir, job, cmd, logger, cloneOpts, operations); err != nil {
			_ = os.RemoveAll(workspaceDir)
			return nil, err
		}
	}

	scriptPaths, err := prepareScripts(ctx, filesStore, job, workspaceDir, logger)
	if err != nil {
		_ = os.RemoveAll(workspaceDir)
		return nil, err
	}

	return &kubernetesWorkspace{
		scriptFilenames: scriptPaths,
		workspaceDir:    workspaceDir,
		logger:          logger,
	}, nil
}

func (w kubernetesWorkspace) Path() string {
	return w.workspaceDir
}

func (w kubernetesWorkspace) WorkingDirectory() string {
	return w.workspaceDir
}

func (w kubernetesWorkspace) ScriptFilenames() []string {
	return w.scriptFilenames
}

func (w kubernetesWorkspace) Remove(ctx context.Context, keepWorkspace bool) {
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

	if w.workspaceDir != "" {
		fmt.Fprintf(handle, "Removing %s\n", w.workspaceDir)
		if rmErr := os.RemoveAll(w.workspaceDir); rmErr != nil {
			fmt.Fprintf(handle, "Operation failed: %s\n", rmErr.Error())
		}
	}
}
