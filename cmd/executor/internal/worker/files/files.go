package files

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ScriptsPath is the location relative to the executor workspace where the executor
// will write scripts required for the execution of the job.
const ScriptsPath = ".sourcegraph-executor"

// Store handles interactions with the file store.
type Store interface {
	// Exists determines if the file exists.
	Exists(ctx context.Context, job types.Job, bucket string, key string) (bool, error)
	// Get retrieves the file.
	Get(ctx context.Context, job types.Job, bucket string, key string) (io.ReadCloser, error)
}

// GetWorkspaceFiles returns the files that should be accessible to jobs within the workspace.
func GetWorkspaceFiles(ctx context.Context, store Store, job types.Job, workingDirectory string) (workspaceFiles []WorkspaceFile, err error) {
	// Construct a map from filenames to file Content that should be accessible to jobs
	// within the workspace. This consists of files supplied within the job record itself,
	// as well as file-version of each script step.
	for relativePath, machineFile := range job.VirtualMachineFiles {
		path, err := filepath.Abs(filepath.Join(workingDirectory, relativePath))
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(path, workingDirectory) {
			return nil, errors.New("refusing to write outside of working directory")
		}
		content, err := getContent(ctx, job, store, machineFile)
		if err != nil {
			return nil, err
		}
		workspaceFiles = append(
			workspaceFiles,
			WorkspaceFile{
				Path:       path,
				Content:    content,
				ModifiedAt: machineFile.ModifiedAt,
			},
		)
	}

	for i, dockerStep := range job.DockerSteps {
		workspaceFiles = append(
			workspaceFiles,
			WorkspaceFile{
				Path:         filepath.Join(workingDirectory, ScriptsPath, ScriptNameFromJobStep(job, i)),
				Content:      []byte(buildScript(dockerStep.Commands)),
				IsStepScript: true,
			},
		)
	}
	return workspaceFiles, nil
}

// WorkspaceFile represents a file that should be accessible to jobs within the workspace.
type WorkspaceFile struct {
	Path         string
	Content      []byte
	ModifiedAt   time.Time
	IsStepScript bool
}

func getContent(ctx context.Context, job types.Job, store Store, machineFile types.VirtualMachineFile) (content []byte, err error) {
	content = machineFile.Content
	if store != nil && machineFile.Bucket != "" && machineFile.Key != "" {
		src, err := store.Get(ctx, job, machineFile.Bucket, machineFile.Key)
		if err != nil {
			return nil, err
		}
		defer src.Close()
		content, err = io.ReadAll(src)
		if err != nil {
			return nil, err
		}
	}
	return content, nil
}

// ScriptPreamble contains a script that checks at runtime if bash is available.
// If it is, we want to be using bash, to support a more natural scripting.
// If not, then we just run with sh.
// This works roughly like the following:
// - If no argument to the script is provided, this is the first run of it. We will use that later to prevent an infinite loop.
// - Determine if a program called bash is on the path
// - If so, we invoke this exact script again, but with the bash on the path, and pass an argument so that this check doesn't happen again.
// - If not, it might be that PATH is not set correctly, but bash is still available at /bin/bash. If that's the case we do the same as above.
// Otherwise we just continue and best effort run the script in sh.
var ScriptPreamble = `
# Only on the first run, check if we can upgrade to bash.
if [ -z "$1" ]; then
  bash_path=$(command -p -v bash)
  set -e
  # Check if bash is present. If so, use bash. Otherwise just keep running with sh.
  if [ -n "$bash_path" ]; then
    exec "${bash_path}" "$0" skip-check
  else
    # If not in the path but still exists at /bin/bash, we can use that.
    if [ -f "/bin/bash" ]; then
      exec /bin/bash "$0" skip-check
    fi
  fi
fi

# Restore default shell behavior.
set +e
# From the actual script, log all commands.
set -x
`

var preambleSlice = []string{ScriptPreamble, ""}

func buildScript(commands []string) string {
	return strings.Join(append(preambleSlice, commands...), "\n") + "\n"
}

// ScriptNameFromJobStep returns the name of the script file for the given job step.
func ScriptNameFromJobStep(job types.Job, i int) string {
	return fmt.Sprintf("%d.%d_%s@%s.sh", job.ID, i, strings.ReplaceAll(job.RepositoryName, "/", "_"), job.Commit)
}
