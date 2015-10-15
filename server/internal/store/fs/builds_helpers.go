package fs

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"strconv"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/rwvfs"
)

// createBuildFile returns a file descriptor pointing to an empty file that stores information
// about the build that was passed. If Attempt == 0, it will create the consequential attempt
// based on the number of files in the directory +1 (excluding "tasks"). If Attempt > 0 it
// will create or truncate that file. The caller is responsible for handling locking mechanisms
// for synchronized access.
func createBuildFile(ctx context.Context, b *sourcegraph.Build) (io.WriteCloser, error) {
	fs := buildStoreVFS(ctx)
	dir := filepath.Join(b.Repo, b.CommitID)
	if err := rwvfs.MkdirAll(fs, dir); err != nil {
		return nil, err
	}
	if b.Attempt == 0 {
		// If no attempt is specified, create the subsequent one.
		fis, err := fs.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		bfis := 0
		for _, fi := range fis {
			if fi.Name() != "tasks" {
				bfis = bfis + 1
			}
		}
		b.Attempt = uint32(bfis + 1)
	}
	fn := filepath.Join(dir, strconv.FormatUint(uint64(b.Attempt), 10))
	return fs.Create(fn)
}

// createTaskFile creates a new empty task file within the sub-folder "tasks" of the repo. It also
// creates child folders which are named based on the Attempt that they belong too. For example,
// task with TaskID 1, belonging to attempt 5 for repo URI my/repo and CommitID a1b2c3 will be
// located in $SGPATH/<buildstore>/my/repo/a1b2c3/tasks/5/1
// As for builds, if a TaskID is not supplied (=0) it will be generated based on the number of
// files in the folder.
func createTaskFile(ctx context.Context, t *sourcegraph.BuildTask) (io.WriteCloser, error) {
	fs := buildStoreVFS(ctx)
	dir := filepath.Join(t.Repo, t.CommitID, "tasks", strconv.FormatUint(uint64(t.Attempt), 10))
	if err := rwvfs.MkdirAll(fs, dir); err != nil {
		return nil, err
	}
	if t.TaskID == 0 {
		fis, err := fs.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		t.TaskID = int64(len(fis) + 1)
	}
	fn := filepath.Join(dir, strconv.FormatInt(t.TaskID, 10))
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

func repoBuildIndexFilename(repo string) string {
	return filepath.Join(repo, buildsIndexFilename)
}

const buildsIndexFilename = "builds-index.json"

// updateRepoBuildIndex adds an entry for the build to the per-repo
// index of all known builds for that repo. Using an index makes it
// much faster to list the builds for a given repo.
//
// TODO(sqs): If there is a race condition, the index can get out of sync with
func updateRepoBuildIndex(ctx context.Context, build sourcegraph.BuildSpec) (err error) {
	builds, err := getRepoBuildIndex(ctx, build.Repo.URI)
	if err != nil {
		return err
	}

	builds = append(builds, build)

	fs := buildStoreVFS(ctx)
	filename := repoBuildIndexFilename(build.Repo.URI)

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
	return json.NewEncoder(f).Encode(builds)
}

func getRepoBuildIndex(ctx context.Context, repo string) ([]sourcegraph.BuildSpec, error) {
	f, err := buildStoreVFS(ctx).Open(repoBuildIndexFilename(repo))
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	var builds []sourcegraph.BuildSpec
	if err = json.NewDecoder(f).Decode(&builds); err != nil {
		return nil, err
	}
	return builds, nil
}

// pathForBuild returns the path for the given build.
func pathForBuild(b sourcegraph.BuildSpec) string {
	return filepath.Join(
		b.Repo.URI,
		b.CommitID,
		strconv.FormatUint(uint64(b.Attempt), 10),
	)
}

func pathForTask(t sourcegraph.TaskSpec) string {
	return filepath.Join(
		t.BuildSpec.Repo.URI,
		t.BuildSpec.CommitID,
		"tasks",
		strconv.FormatUint(uint64(t.BuildSpec.Attempt), 10),
		strconv.FormatInt(int64(t.TaskID), 10),
	)
}
