package git

import (
	"context"
	"fmt"
	"io/fs"
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

	cmd := gitserver.NewClient(db).Command("git", args...)
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
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
	cmd := gitserver.NewClient(db).Command("git", "ls-tree", "--name-only", "-r", string(commit), "--")
	cmd.Repo = repo

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

// DevNullSHA 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t
// tree /dev/null`, which is used as the base when computing the `git diff` of
// the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
