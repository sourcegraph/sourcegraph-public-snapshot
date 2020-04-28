package gitserver

import (
	"bytes"
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

// DirectoryChildren determines all children known to git for the given directory names via an invocation
// of git ls-tree. The keys of the resulting map are the input (unsanitized) dirnames, and the value of
// that key are the files nested under that directory.
func DirectoryChildren(db db.DB, repositoryID int, commit string, dirnames []string) (map[string][]string, error) {
	// TODO(efritz) - remove dependency on codeintel/db package
	repoName, err := db.RepoName(context.Background(), repositoryID)
	if err != nil {
		return nil, err
	}

	cmd := gitserver.DefaultClient.Command("git", append([]string{"ls-tree", "--name-only", commit, "--"}, cleanDirectoriesForLsTree(dirnames)...)...)
	cmd.Repo = gitserver.Repo{Name: api.RepoName(repoName)}
	out, err := cmd.CombinedOutput(context.Background())
	if err != nil {
		return nil, err
	}

	return parseDirectoryChildren(dirnames, strings.Split(string(bytes.TrimSpace(out)), "\n")), nil
}

// cleanDirectoriesForLsTree sanitizes the input dirnames to a git ls-tree command. There are a
// few pecularities handled here:
//
//   1. The root of the tree must be indicated with `.`, and
//   2. In order for git ls-tree to return a directory's contents, the name must end in a slash.
func cleanDirectoriesForLsTree(dirnames []string) []string {
	var args []string
	for _, dir := range dirnames {
		if dir == "" {
			args = append(args, ".")
		} else {
			if !strings.HasSuffix(dir, "/") {
				dir += "/"
			}
			args = append(args, dir)
		}
	}

	return args
}

// parseDirectoryChildren converts the flat list of files from git ls-tree into a map. The keys of the
// resulting map are the input (unsanitized) dirnames, and the value of that key are the files nested
// under that directory.
func parseDirectoryChildren(dirnames []string, paths []string) map[string][]string {
	childrenMap := map[string][]string{}
	for _, dir := range dirnames {
		if dir == "" {
			var children []string
			for _, path := range paths {
				if !strings.Contains(path, "/") {
					children = append(children, path)
				}
			}

			childrenMap[dir] = children
		} else {
			var children []string
			for _, path := range paths {
				if strings.HasPrefix(path, dir) {
					children = append(children, path)
				}
			}

			childrenMap[dir] = children
		}
	}

	return childrenMap
}
