package apidocs

import (
	"path"
	"strings"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// defsForDir returns all of the definitions found in every file of the given
// directory. There is potential for a perf improvement here by instead adding
// a specific flag for this in DefListOptions.
func defsForDir(ctx context.Context, rev sourcegraph.RepoRevSpec, dir string) (defs []*sourcegraph.Def, err error) {
	cl := sourcegraph.NewClientFromContext(ctx)

	// Resolve the rev if not already, this is required for Defs.List below to
	// succeed.
	if err := resolveRevSpec(ctx, &rev); err != nil {
		return nil, err
	}

	// TODO(slimsag): We only need files from a specific directory, but List
	// returns all files in the entire repository. Room for optimization here too.
	list, err := cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{
		Rev: rev,
	})
	if err != nil {
		return nil, err
	}

	// Now for each file in the directory, list it's defs.
	var files []string
	for _, f := range list.Files {
		// Check that the file is in the requested directory.
		if path.Dir(f) != dir {
			continue
		}
		files = append(files, f)
	}
	filesDefs, err := cl.Defs.List(ctx, &sourcegraph.DefListOptions{
		RepoRevs:    []string{rev.RepoSpec.SpecString() + "@" + rev.CommitID},
		Exported:    true,
		Files:       files,
		Doc:         true,
		ListOptions: sourcegraph.ListOptions{PerPage: 10000},
	})
	if err != nil {
		return nil, err
	}
	return filesDefs.Defs, nil
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

// resolveRevSpec resolves the RepoRevSpec if it is not already.
func resolveRevSpec(ctx context.Context, rev *sourcegraph.RepoRevSpec) error {
	cl := sourcegraph.NewClientFromContext(ctx)
	if rev.Rev == "" {
		// Determine default branch.
		repo, err := cl.Repos.Get(ctx, &rev.RepoSpec)
		if err != nil {
			return err
		}
		rev.Rev = repo.DefaultBranch
	}
	if !rev.Resolved() {
		commit, err := cl.Repos.GetCommit(ctx, rev)
		if err != nil {
			return err
		}
		rev.CommitID = string(commit.ID)
	}
	return nil
}
