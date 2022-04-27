package git

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/vcs/util"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Stat returns a FileInfo describing the named file at commit.
func Stat(ctx context.Context, db database.DB, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
	if Mocks.Stat != nil {
		return Mocks.Stat(commit, path)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: Stat")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	path = util.Rel(path)

	fi, err := gitserver.LStat(ctx, db, checker, repo, commit, path)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

// LsFiles returns the output of `git ls-files`
func LsFiles(ctx context.Context, db database.DB, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, pathspecs ...gitserver.Pathspec) ([]string, error) {
	if Mocks.LsFiles != nil {
		return Mocks.LsFiles(repo, commit)
	}
	args := []string{
		"ls-files",
		"-z",
		"--with-tree",
		string(commit),
	}

	if len(pathspecs) > 0 {
		args = append(args, "--")
		for _, pathspec := range pathspecs {
			args = append(args, string(pathspec))
		}
	}

	cmd := gitserver.NewClient(db).GitCommand(repo, args...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}

	files := strings.Split(string(out), "\x00")
	// Drop trailing empty string
	if len(files) > 0 && files[len(files)-1] == "" {
		files = files[:len(files)-1]
	}
	return filterPaths(ctx, repo, checker, files)
}

// ListFiles returns a list of root-relative file paths matching the given
// pattern in a particular commit of a repository.
func ListFiles(ctx context.Context, db database.DB, repo api.RepoName, commit api.CommitID, pattern *regexp.Regexp, checker authz.SubRepoPermissionChecker) (_ []string, err error) {
	cmd := gitserver.NewClient(db).GitCommand(repo, "ls-tree", "--name-only", "-r", string(commit), "--")

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	var matching []string
	for _, path := range strings.Split(string(out), "\n") {
		if pattern.MatchString(path) {
			matching = append(matching, path)
		}
	}

	return filterPaths(ctx, repo, checker, matching)
}

// 🚨 SECURITY: All git methods that deal with file or path access need to have
// sub-repo permissions applied
func filterPaths(ctx context.Context, repo api.RepoName, checker authz.SubRepoPermissionChecker, paths []string) ([]string, error) {
	if !authz.SubRepoEnabled(checker) {
		return paths, nil
	}
	a := actor.FromContext(ctx)
	filtered, err := authz.FilterActorPaths(ctx, checker, a, repo, paths)
	if err != nil {
		return nil, errors.Wrap(err, "filtering paths")
	}
	return filtered, nil
}

// ListDirectoryChildren fetches the list of children under the given directory
// names. The result is a map keyed by the directory names with the list of files
// under each.
func ListDirectoryChildren(
	ctx context.Context,
	db database.DB,
	checker authz.SubRepoPermissionChecker,
	repo api.RepoName,
	commit api.CommitID,
	dirnames []string,
) (map[string][]string, error) {
	args := []string{"ls-tree", "--name-only", string(commit), "--"}
	args = append(args, cleanDirectoriesForLsTree(dirnames)...)
	cmd := gitserver.NewClient(db).GitCommand(repo, args...)

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	paths := strings.Split(string(out), "\n")
	if authz.SubRepoEnabled(checker) {
		paths, err = authz.FilterActorPaths(ctx, checker, actor.FromContext(ctx), repo, paths)
		if err != nil {
			return nil, err
		}
	}
	return parseDirectoryChildren(dirnames, paths), nil
}

// cleanDirectoriesForLsTree sanitizes the input dirnames to a git ls-tree command. There are a
// few peculiarities handled here:
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

// DevNullSHA 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t
// tree /dev/null`, which is used as the base when computing the `git diff` of
// the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
