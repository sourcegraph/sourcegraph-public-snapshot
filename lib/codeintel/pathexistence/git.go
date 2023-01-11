package pathexistence

import (
	"context"
	"os/exec"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitFunc func(args ...string) (string, error)

// GitGetChildren lists all the children under the givem directories for the given commit.
//
// NOTE: A copy of this function was added to
// sourcegraph/sourcegraph/internal/vcs/git called ListDirectoryChildren as we
// don't want to rely on this package from there.
func GitGetChildren(gitFunc GitFunc, commit string, dirnames []string) (map[string][]string, error) {
	out, err := gitFunc(
		append(
			[]string{"ls-tree", "--name-only", commit, "--"},
			cleanDirectoriesForLsTree(dirnames)...,
		)...,
	)

	if err != nil {
		return nil, errors.Wrap(err, "Running ls-tree")
	}

	return parseDirectoryChildren(dirnames, strings.Split(out, "\n")), nil
}

func LocalGitGetChildrenFunc(repoRoot string) GetChildrenFunc {
	return func(ctx context.Context, dirnames []string) (map[string][]string, error) {
		return GitGetChildren(
			func(args ...string) (string, error) {
				out, err := exec.Command(
					"git",
					append(
						[]string{"-C", repoRoot},
						args...,
					)...,
				).CombinedOutput()

				return string(out), err
			},
			"HEAD",
			dirnames,
		)
	}
}

// cleanDirectoriesForLsTree sanitizes the input dirnames to a git ls-tree command. There are a
// few peculiarities handled here:
//
//  1. The root of the tree must be indicated with `.`, and
//  2. In order for git ls-tree to return a directory's contents, the name must end in a slash.
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
// under that directory. If dirnames contains a directory that encloses another, then the paths will
// be placed into the key sharing the longest path prefix.
func parseDirectoryChildren(dirnames, paths []string) map[string][]string {
	childrenMap := map[string][]string{}

	// Ensure each directory has an entry, even if it has no children
	// listed in the gitserver output.
	for _, dirname := range dirnames {
		childrenMap[dirname] = nil
	}

	// Order directory names by length (biggest first) so that we assign
	// paths to the most specific enclosing directory in the following loop.
	sort.Slice(dirnames, func(i, j int) bool {
		return len(dirnames[i]) > len(dirnames[j])
	})

	for _, path := range paths {
		if strings.Contains(path, "/") {
			for _, dirname := range dirnames {
				if strings.HasPrefix(path, dirname) {
					childrenMap[dirname] = append(childrenMap[dirname], path)
					break
				}
			}
		} else if len(dirnames) > 0 && dirnames[len(dirnames)-1] == "" {
			// No need to loop here. If we have a root input directory it
			// will necessarily be the last element due to the previous
			// sorting step.
			childrenMap[""] = append(childrenMap[""], path)
		}
	}

	return childrenMap
}
