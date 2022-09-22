package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func prepareScripts(
	ctx context.Context,
	job executor.Job,
	workspaceDir string,
	commandRunner command.Runner,
	commandLogger command.Logger,
) ([]string, error) {
	// Create the scripts path.
	if err := os.MkdirAll(filepath.Join(workspaceDir, command.ScriptsPath), os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "creating script path")
	}

	// Construct a map from filenames to file content that should be accessible to jobs
	// within the workspace. This consists of files supplied within the job record itself,
	// as well as file-version of each script step.
	workspaceFileContentsByPath := map[string][]byte{}

	for relativePath, content := range job.VirtualMachineFiles {
		path, err := filepath.Abs(filepath.Join(workspaceDir, relativePath))
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(path, workspaceDir) {
			return nil, errors.Errorf("refusing to write outside of working directory")
		}

		workspaceFileContentsByPath[path] = []byte(content)
	}

	scriptNames := make([]string, 0, len(job.DockerSteps))
	for i, dockerStep := range job.DockerSteps {
		scriptName := scriptNameFromJobStep(job, i)
		scriptNames = append(scriptNames, scriptName)

		path := filepath.Join(workspaceDir, command.ScriptsPath, scriptName)
		workspaceFileContentsByPath[path] = buildScript(dockerStep)
	}

	if err := writeFiles(workspaceFileContentsByPath, commandLogger); err != nil {
		return nil, errors.Wrap(err, "failed to write virtual machine files")
	}

	return scriptNames, nil
}

var scriptPreamble = `
set -x
`

func buildScript(dockerStep executor.DockerStep) []byte {
	return []byte(strings.Join(append([]string{scriptPreamble, ""}, dockerStep.Commands...), "\n") + "\n")
}

func scriptNameFromJobStep(job executor.Job, i int) string {
	return fmt.Sprintf("%d.%d_%s@%s.sh", job.ID, i, strings.ReplaceAll(job.RepositoryName, "/", "_"), job.Commit)
}

// writeFiles writes the content of the given map to the filesystem.
func writeFiles(workspaceFileContentsByPath map[string][]byte, logger command.Logger) (err error) {
	// Bail out early if nothing to do, we don't want to spawn an empty log group.
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

	for path, content := range workspaceFileContentsByPath {
		// Ensure the path exists.
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(path, content, os.ModePerm); err != nil {
			return err
		}

		fmt.Fprintf(handle, "Wrote %s\n", path)
	}

	return nil
}
