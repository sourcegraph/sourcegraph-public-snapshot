package workspace

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/files"
	"github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func prepareScripts(
	ctx context.Context,
	filesStore files.Store,
	job types.Job,
	workspaceDir string,
	commandLogger cmdlogger.Logger,
) ([]string, error) {
	// Create the scripts path.
	if err := os.MkdirAll(filepath.Join(workspaceDir, files.ScriptsPath), os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "creating script path")
	}

	workspaceFiles, err := files.GetWorkspaceFiles(ctx, filesStore, job, workspaceDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get workspace files")
	}

	if err = writeFiles(commandLogger, workspaceFiles); err != nil {
		return nil, errors.Wrap(err, "failed to write virtual machine files")
	}

	scriptNames := make([]string, 0, len(job.DockerSteps))
	for _, file := range workspaceFiles {
		if file.IsStepScript {
			scriptNames = append(scriptNames, filepath.Base(file.Path))
		}
	}

	return scriptNames, nil
}

// writeFiles writes to the filesystem the content in the given map.
func writeFiles(logger cmdlogger.Logger, workspaceFiles []files.WorkspaceFile) (err error) {
	// Bail out early if nothing to do, we don't need to spawn an empty log group.
	if len(workspaceFiles) == 0 {
		return nil
	}

	handle := logger.LogEntry("setup.fs.extras", nil)
	defer func() {
		if err == nil {
			handle.Finalize(0)
		} else {
			handle.Finalize(1)
		}

		_ = handle.Close()
	}()

	for _, wf := range workspaceFiles {
		// Ensure the path exists.
		if err := os.MkdirAll(filepath.Dir(wf.Path), os.ModePerm); err != nil {
			return err
		}

		var src io.ReadCloser

		// Log how long it takes to write the files
		start := time.Now()
		src = io.NopCloser(bytes.NewReader(wf.Content))

		f, err := os.Create(wf.Path)
		if err != nil {
			return err
		}

		if _, err = io.Copy(f, src); err != nil {
			return errors.Append(err, f.Close())
		}

		if err = f.Close(); err != nil {
			return err
		}

		// Ensure the file has permissions to be run
		if err = os.Chmod(wf.Path, os.ModePerm); err != nil {
			return err
		}

		// Set modified time for caching (if provided)
		if !wf.ModifiedAt.IsZero() {
			if err = os.Chtimes(wf.Path, wf.ModifiedAt, wf.ModifiedAt); err != nil {
				return err
			}
		}

		if _, err = handle.Write([]byte(fmt.Sprintf("Wrote %s in %s\n", wf.Path, time.Since(start)))); err != nil {
			return err
		}
	}

	return nil
}
