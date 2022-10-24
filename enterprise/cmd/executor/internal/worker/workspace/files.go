package workspace

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func prepareScripts(
	ctx context.Context,
	filesStore store.FilesStore,
	job executor.Job,
	workspaceDir string,
	commandLogger command.Logger,
) ([]string, error) {
	// Create the scripts path.
	if err := os.MkdirAll(filepath.Join(workspaceDir, command.ScriptsPath), os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "creating script path")
	}

	// Construct a map from filenames to file content that should be accessible to jobs
	// within the workspace. This consists of files supplied within the job record itself,
	// as well as file-version of each script step.
	workspaceFilesByPath := map[string]workspaceFile{}

	for relativePath, machineFile := range job.VirtualMachineFiles {
		path, err := filepath.Abs(filepath.Join(workspaceDir, relativePath))
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(path, workspaceDir) {
			return nil, errors.Errorf("refusing to write outside of working directory")
		}
		// Either write raw content that has already been provided or retrieve it from the store.
		workspaceFilesByPath[path] = workspaceFile{
			content:    []byte(machineFile.Content),
			bucket:     machineFile.Bucket,
			key:        machineFile.Key,
			modifiedAt: machineFile.ModifiedAt,
		}
	}

	scriptNames := make([]string, 0, len(job.DockerSteps))
	for i, dockerStep := range job.DockerSteps {
		scriptName := scriptNameFromJobStep(job, i)
		scriptNames = append(scriptNames, scriptName)

		path := filepath.Join(workspaceDir, command.ScriptsPath, scriptName)
		workspaceFilesByPath[path] = buildScript(dockerStep)
	}

	if err := writeFiles(ctx, filesStore, workspaceFilesByPath, commandLogger); err != nil {
		return nil, errors.Wrap(err, "failed to write virtual machine files")
	}

	return scriptNames, nil
}

type workspaceFile struct {
	content    []byte
	bucket     string
	key        string
	modifiedAt time.Time
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

func buildScript(dockerStep executor.DockerStep) workspaceFile {
	return workspaceFile{content: []byte(strings.Join(append([]string{ScriptPreamble, ""}, dockerStep.Commands...), "\n") + "\n")}
}

func scriptNameFromJobStep(job executor.Job, i int) string {
	return fmt.Sprintf("%d.%d_%s@%s.sh", job.ID, i, strings.ReplaceAll(job.RepositoryName, "/", "_"), job.Commit)
}

// writeFiles writes to the filesystem the content in the given map.
func writeFiles(ctx context.Context, store store.FilesStore, workspaceFileContentsByPath map[string]workspaceFile, logger command.Logger) (err error) {
	// Bail out early if nothing to do, we don't need to spawn an empty log group.
	if len(workspaceFileContentsByPath) == 0 {
		return nil
	}

	handle := logger.Log("setup.fs.extras", nil)
	defer func() {
		if err == nil {
			handle.Finalize(0)
		} else {
			handle.Finalize(1)
		}

		handle.Close()
	}()

	for path, wf := range workspaceFileContentsByPath {
		// Ensure the path exists.
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		var src io.ReadCloser

		// Log how long it takes to write the files
		start := time.Now()
		if store != nil && wf.bucket != "" && wf.key != "" {
			src, err = store.Get(ctx, wf.bucket, wf.key)
			if err != nil {
				return err
			}
		} else {
			src = io.NopCloser(bytes.NewReader(wf.content))
		}

		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = io.Copy(f, src); err != nil {
			return err
		}

		// Ensure the file has permissions to be run
		if err = os.Chmod(path, os.ModePerm); err != nil {
			return err
		}

		// Set modified time for caching (if provided)
		if !wf.modifiedAt.IsZero() {
			if err = os.Chtimes(path, wf.modifiedAt, wf.modifiedAt); err != nil {
				return err
			}
		}

		handle.Write([]byte(fmt.Sprintf("Wrote %s in %s\n", path, time.Since(start))))
	}

	return nil
}
