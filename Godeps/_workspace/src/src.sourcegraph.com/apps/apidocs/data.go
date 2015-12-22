package apidocs

import (
	"path"
	"strings"
	"sync"

	"code.google.com/p/rog-go/parallel"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// defsForDir returns all of the definitions found in every file of the given
// directory. There is potential for a perf improvement here by instead adding
// a specific flag for this in DefListOptions.
func defsForDir(ctx context.Context, rev sourcegraph.RepoRevSpec, dir string) (defs []*sourcegraph.Def, err error) {
	cl := sourcegraph.NewClientFromContext(ctx)

	// TODO(slimsag): We only need files from a specific directory, but List
	// returns all files in the entire repository. Room for optimization here too.
	list, err := cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{
		Rev: rev,
	})
	if err != nil {
		return nil, err
	}

	// Now for each file in the directory, list it's defs and accumulate them. We
	// do this in parallel.
	run := parallel.NewRun(8)
	appension := &sync.Mutex{}
	for _, f := range list.Files {
		// Check that the file is in the requested directory.
		if path.Dir(f) != dir {
			continue
		}

		f := f // copy to avoid data race
		run.Do(func() error {
			fileDefs, err := cl.Defs.List(ctx, &sourcegraph.DefListOptions{
				RepoRevs:    []string{rev.RepoSpec.SpecString()},
				Exported:    true,
				File:        f,
				Doc:         true,
				ListOptions: sourcegraph.ListOptions{PerPage: 10000},
			})
			if err != nil {
				return err
			}

			// Append to defs list.
			appension.Lock()
			defs = append(defs, fileDefs.Defs...)
			appension.Unlock()
			return nil
		})
	}
	if err := run.Wait(); err != nil {
		return nil, err
	}
	return defs, nil
}

// subDirsForDir returns the subdirectores for the given directory.
func subDirsForDir(ctx context.Context, rev sourcegraph.RepoRevSpec, requestedDir string) (dirs []string, err error) {
	cl := sourcegraph.NewClientFromContext(ctx)

	// TODO(slimsag): We only need directories that are under a specific path
	// prefix, but List doesn't allow this. Room for optimization here.
	list, err := cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{
		Rev: rev,
	})
	if err != nil {
		return nil, err
	}

	if requestedDir == "." {
		requestedDir = ""
	}

	dedupDirs := make(map[string]struct{})
	requestedDirElems := countElements(requestedDir)
	for _, f := range list.Files {
		if !strings.HasPrefix(f, requestedDir) {
			continue
		}

		dir := path.Dir(f)
		if dir == "." {
			continue
		}
		subSplit := strings.Split(dir, "/")
		if len(subSplit) < requestedDirElems+1 {
			continue
		}

		dir = strings.Join(subSplit[:requestedDirElems+1], "/")
		if _, ok := dedupDirs[dir]; ok {
			continue
		}
		dedupDirs[dir] = struct{}{}
		dirs = append(dirs, dir)
	}

	return dirs, nil
}

// countElems counts the number of elements in a path string.
func countElements(path string) (count int) {
	for _, p := range strings.Split(path, "/") {
		if strings.TrimSpace(p) != "" {
			count++
		}
	}
	return
}
