package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper/log"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/batcheshelper/util"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Post processes the workspace after the Batch Change step.
func Post(
	ctx context.Context,
	logger *log.Logger,
	stepIdx int,
	executionInput batcheslib.WorkspacesExecutionInput,
	previousResult execution.AfterStepResult,
	workspaceFilesPath string,
) error {
	// Sometimes the files belong to different users. Mark the repository directory as safe.
	if _, err := runGitCmd(ctx, "config", "--global", "--add", "safe.directory", "/data/repository"); err != nil {
		return errors.Wrap(err, "failed to mark repository directory as safe")
	}

	// Generate the diff.
	if _, err := runGitCmd(ctx, "add", "--all"); err != nil {
		return errors.Wrap(err, "failed to add all files to git")
	}
	diff, err := runGitCmd(ctx, "diff", "--cached", "--no-prefix", "--binary")
	if err != nil {
		return errors.Wrap(err, "failed to generate diff")
	}

	// Read the stdout of the current step.
	stdout, err := os.ReadFile(fmt.Sprintf("stdout%d.log", stepIdx))
	if err != nil {
		return errors.Wrap(err, "failed to read stdout file")
	}

	// Read the stderr of the current step.
	stderr, err := os.ReadFile(fmt.Sprintf("stderr%d.log", stepIdx))
	if err != nil {
		return errors.Wrap(err, "failed to read stderr file")
	}

	// Build the step result.
	stepResult := execution.AfterStepResult{
		Version:   2,
		Stdout:    string(stdout),
		Stderr:    string(stderr),
		StepIndex: stepIdx,
		Diff:      diff,
		// Those will be set below.
		Outputs: make(map[string]interface{}),
	}

	// Render the step outputs.
	changes, err := git.ChangesInDiff(previousResult.Diff)
	if err != nil {
		return errors.Wrap(err, "failed to get changes in diff")
	}
	outputs := previousResult.Outputs
	stepContext := template.StepContext{
		BatchChange: executionInput.BatchChangeAttributes,
		Repository: template.Repository{
			Name:        executionInput.Repository.Name,
			Branch:      executionInput.Branch.Name,
			FileMatches: executionInput.SearchResultPaths,
		},
		Outputs: outputs,
		Steps: template.StepsContext{
			Path:    executionInput.Path,
			Changes: changes,
		},
		PreviousStep: previousResult,
		Step:         stepResult,
	}

	// Render and evaluate outputs.
	step := executionInput.Steps[stepIdx]
	if err = batcheslib.SetOutputs(step.Outputs, outputs, &stepContext); err != nil {
		return errors.Wrap(err, "setting outputs")
	}
	for k, v := range outputs {
		stepResult.Outputs[k] = v
	}

	// Serialize the step result to disk.
	stepResultBytes, err := json.Marshal(stepResult)
	if err != nil {
		return errors.Wrap(err, "marshalling step result")
	}
	if err = os.WriteFile(fmt.Sprintf("step%d.json", stepIdx), stepResultBytes, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to write step result file")
	}

	// Build and write the cache key
	key := cache.KeyForWorkspace(
		&executionInput.BatchChangeAttributes,
		batcheslib.Repository{
			ID:          executionInput.Repository.ID,
			Name:        executionInput.Repository.Name,
			BaseRef:     executionInput.Branch.Name,
			BaseRev:     executionInput.Branch.Target.OID,
			FileMatches: executionInput.SearchResultPaths,
		},
		executionInput.Path,
		os.Environ(),
		executionInput.OnlyFetchWorkspace,
		executionInput.Steps,
		stepIdx,
		fileMetadataRetriever{workingDirectory: workspaceFilesPath},
	)

	k, err := key.Key()
	if err != nil {
		return errors.Wrap(err, "failed to compute cache key")
	}

	err = logger.WriteEvent(
		batcheslib.LogEventOperationCacheAfterStepResult,
		batcheslib.LogEventStatusSuccess,
		&batcheslib.CacheAfterStepResultMetadata{Key: k, Value: stepResult},
	)
	if err != nil {
		return err
	}

	// Cleanup the workspace.
	return cleanupWorkspace(stepIdx, workspaceFilesPath)
}

func runGitCmd(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = "repository"

	return cmd.Output()
}

type fileMetadataRetriever struct {
	workingDirectory string
}

var _ cache.MetadataRetriever = fileMetadataRetriever{}

func (f fileMetadataRetriever) Get(steps []batcheslib.Step) ([]cache.MountMetadata, error) {
	var mountsMetadata []cache.MountMetadata
	for _, step := range steps {
		// Build up the metadata for each mount for each step
		for _, mount := range step.Mount {
			metadata, err := f.getMountMetadata(f.workingDirectory, mount.Path)
			if err != nil {
				return nil, err
			}
			// A mount could be a directory containing multiple files
			mountsMetadata = append(mountsMetadata, metadata...)
		}
	}
	return mountsMetadata, nil
}

func (f fileMetadataRetriever) getMountMetadata(baseDir string, path string) ([]cache.MountMetadata, error) {
	fullPath := path
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(baseDir, path)
	}
	info, err := os.Stat(fullPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.Newf("path %s does not exist", path)
	} else if err != nil {
		return nil, err
	}
	var metadata []cache.MountMetadata
	if info.IsDir() {
		dirMetadata, err := f.getDirectoryMountMetadata(fullPath)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, dirMetadata...)
	} else {
		relativePath, err := filepath.Rel(f.workingDirectory, fullPath)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, cache.MountMetadata{Path: relativePath, Size: info.Size(), Modified: info.ModTime().UTC()})
	}
	return metadata, nil
}

// getDirectoryMountMetadata reads all the files in the directory with the given
// path and returns the cache.MountMetadata for all of them.
func (f fileMetadataRetriever) getDirectoryMountMetadata(path string) ([]cache.MountMetadata, error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var metadata []cache.MountMetadata
	for _, dirEntry := range dir {
		// Go back to the very start. Need to get the FileInfo again for the new path and figure out if it is a
		// directory or a file.
		fileMetadata, err := f.getMountMetadata(path, dirEntry.Name())
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, fileMetadata...)
	}
	return metadata, nil
}

func cleanupWorkspace(step int, workspaceFilesPath string) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "getting working directory")
	}
	tmpFileDir := util.FilesMountPath(wd, step)
	if err = os.RemoveAll(tmpFileDir); err != nil {
		return errors.Wrap(err, "removing files mount")
	}
	return os.RemoveAll(workspaceFilesPath)
}
