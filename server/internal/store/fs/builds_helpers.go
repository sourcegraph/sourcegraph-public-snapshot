package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// createBuildFile returns a file descriptor pointing to an empty file
// that stores information about the build that was passed. If ID ==
// 0, it will create the new build based on the number of files in the
// directory +1 (excluding "tasks"). If ID > 0 it will create or
// truncate that file. The caller is responsible for handling locking
// mechanisms for synchronized access.
func createBuildFile(ctx context.Context, b *sourcegraph.Build) (io.WriteCloser, error) {
	fs := buildStoreVFS(ctx)

	// This dir path must stay consistent with the dirForBuild logic.
	dir := b.Repo
	if err := rwvfs.MkdirAll(fs, dir); err != nil {
		return nil, err
	}
	if b.ID == 0 {
		// If no attempt is specified, create the subsequent one.
		fis, err := fs.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		bfis := 0
		for _, fi := range fis {
			if fi.Mode().IsDir() {
				bfis++
			}
		}
		b.ID = uint64(bfis + 1)
	}
	fn := filepath.Join(dir, fmt.Sprint(b.ID), "build.json")
	if err := rwvfs.MkdirAll(fs, filepath.Dir(fn)); err != nil {
		return nil, err
	}
	return fs.Create(fn)
}

// createTaskFile creates a new empty task file within the sub-folder
// "tasks" of the repo. For example, task with ID 1, belonging to
// build ID 5 for repo URI my/repo will be located in
// $SGPATH/<buildstore>/my/repo/5/tasks/1 If a task ID is not supplied
// (=0) it will be generated based on the number of files in the
// folder.
func createTaskFile(ctx context.Context, t *sourcegraph.BuildTask) (io.WriteCloser, error) {
	fs := buildStoreVFS(ctx)
	dir := filepath.Join(dirForBuild(t.Build), "tasks")
	if err := rwvfs.MkdirAll(fs, dir); err != nil {
		return nil, err
	}
	if t.ID == 0 {
		fis, err := fs.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		t.ID = uint64(len(fis) + 1)
	}
	fn := filepath.Join(dir, fmt.Sprint(t.ID)+".json")
	return fs.Create(fn)
}

// getQueue opens (or creates) the given filename and attempts to decode JSON
// into v.
func getQueue(ctx context.Context, fileName string, v interface{}) error {
	fs := buildStoreVFS(ctx)
	stat, err := fs.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if stat.Size() == 0 {
		// no queue
		return nil
	}
	f, err := fs.Open(fileName)
	if err != nil {
		return err
	}
	defer func() {
		var errors MultiError
		err = errors.verify(f.Close(), err)
	}()
	if err = json.NewDecoder(f).Decode(v); err != nil {
		return err
	}
	return nil
}

// replaceQueue truncates the given filename and attempts to JSON encode v into
// it.
func replaceQueue(ctx context.Context, fileName string, v interface{}) error {
	f, err := buildStoreVFS(ctx).Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		var errors MultiError
		err = errors.verify(f.Close(), err)
	}()
	if err = json.NewEncoder(f).Encode(v); err != nil {
		return err
	}
	return nil
}

// createQueueEntry adds the passed Build to the queue. The caller is responsible
// for handling locking mechanisms for synchronized access.
func createBuildQueueEntry(ctx context.Context, b sourcegraph.Build) error {
	var queue []sourcegraph.BuildSpec
	if err := getQueue(ctx, buildQueueFilename, &queue); err != nil {
		return err
	}
	queue = append(queue, b.Spec())
	return replaceQueue(ctx, buildQueueFilename, queue)
}

func repoBuildCommitIDIndexFilename(repo string) string {
	return filepath.Join(repo, buildsCommitIDIndexFilename)
}

const buildsCommitIDIndexFilename = "builds-commit-id-index.json"

// updateRepoBuildIndex adds an entry for the build to the per-repo
// index of all known builds for that repo. Using an index makes it
// much faster to list the builds for a given repo.
//
// TODO(sqs): If there is a race condition, the index can get out of sync with
func updateRepoBuildCommitIDIndex(ctx context.Context, build sourcegraph.BuildSpec, commitID string) (err error) {
	idx, err := getRepoBuildIndex(ctx, build.Repo.URI)
	if err != nil {
		return err
	}
	if idx == nil {
		idx = repoBuildCommitIDIndex{}
	}

	idx[commitID] = append(idx[commitID], build)

	fs := buildStoreVFS(ctx)
	filename := repoBuildCommitIDIndexFilename(build.Repo.URI)

	if err := rwvfs.MkdirAll(fs, filepath.Dir(filename)); err != nil {
		return err
	}

	f, err := fs.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()
	return json.NewEncoder(f).Encode(idx)
}

type repoBuildCommitIDIndex map[string][]sourcegraph.BuildSpec

func getRepoBuildIndex(ctx context.Context, repo string) (repoBuildCommitIDIndex, error) {
	f, err := buildStoreVFS(ctx).Open(repoBuildCommitIDIndexFilename(repo))
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	var idx repoBuildCommitIDIndex
	if err := json.NewDecoder(f).Decode(&idx); err != nil {
		return nil, err
	}
	return idx, nil
}

// dirForBuild returns the dir for the given build.
func dirForBuild(b sourcegraph.BuildSpec) string {
	return filepath.Join(b.Repo.URI, fmt.Sprint(b.ID))
}

func filenameForTask(t sourcegraph.TaskSpec) string {
	return filepath.Join(dirForBuild(t.Build), "tasks", fmt.Sprint(t.ID)+".json")
}
