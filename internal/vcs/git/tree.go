package git

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	stdlibpath "path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/golang/groupcache/lru"
	"gopkg.in/src-d/go-git.v4/plumbing/format/config"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/vcs/util"
)

// Stat returns a FileInfo describing the named file at commit.
func Stat(ctx context.Context, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
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

	fi, err := lStat(ctx, repo, commit, path)
	if err != nil {
		return nil, err
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		// Deref symlink.
		b, err := readFileBytes(ctx, repo, commit, path, 0)
		if err != nil {
			return nil, err
		}
		// Resolve relative links from the directory path is in
		symlink := filepath.Join(filepath.Dir(path), string(b))
		fi2, err := lStat(ctx, repo, commit, symlink)
		if err != nil {
			return nil, err
		}
		fi2.(*util.FileInfo).Name_ = fi.Name()
		return fi2, nil
	}

	return fi, nil
}

// ReadDir reads the contents of the named directory at commit.
func ReadDir(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
	if Mocks.ReadDir != nil {
		return Mocks.ReadDir(commit, path, recurse)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: ReadDir")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	span.SetTag("Recurse", recurse)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	if path != "" {
		// Trailing slash is necessary to ls-tree under the dir (not just
		// to list the dir's tree entry in its parent dir).
		path = filepath.Clean(util.Rel(path)) + "/"
	}
	return lsTree(ctx, repo, commit, path, recurse)
}

// LsFiles returns the output of `git ls-files`
func LsFiles(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID) ([]string, error) {
	if Mocks.LsFiles != nil {
		return Mocks.LsFiles(repo, commit)
	}
	args := []string{
		"ls-files",
		"-z",
		"--with-tree",
		string(commit),
	}
	cmd := gitserver.DefaultClient.Command("git", args...)
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

// lStat returns a FileInfo describing the named file at commit. If the file is a symbolic link, the
// returned FileInfo describes the symbolic link.  lStat makes no attempt to follow the link.
func lStat(ctx context.Context, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: lStat")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	path = filepath.Clean(util.Rel(path))

	if path == "." {
		// Special case root, which is not returned by `git ls-tree`.
		obj, err := gitserver.DefaultClient.GetObject(ctx, repo, string(commit)+"^{tree}")
		if err != nil {
			return nil, err
		}
		return &util.FileInfo{Mode_: os.ModeDir, Sys_: objectInfo(obj.ID)}, nil
	}

	fis, err := lsTree(ctx, repo, commit, path, false)
	if err != nil {
		return nil, err
	}
	if len(fis) == 0 {
		return nil, &os.PathError{Op: "ls-tree", Path: path, Err: os.ErrNotExist}
	}

	return fis[0], nil
}

// lsTreeRootCache caches the result of running `git ls-tree ...` on a repository's root path
// (because non-root paths are likely to have a lower cache hit rate). It is intended to improve the
// perceived performance of large monorepos, where the tree for a given repo+commit (usually the
// repo's latest commit on default branch) will be requested frequently and would take multiple
// seconds to compute if uncached.
var (
	lsTreeRootCacheMu sync.Mutex
	lsTreeRootCache   = lru.New(5)
)

// lsTree returns ls of tree at path.
func lsTree(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
	if path != "" || !recurse {
		// Only cache the root recursive ls-tree.
		return lsTreeUncached(ctx, repo, commit, path, recurse)
	}

	key := string(repo) + ":" + string(commit) + ":" + path
	lsTreeRootCacheMu.Lock()
	v, ok := lsTreeRootCache.Get(key)
	lsTreeRootCacheMu.Unlock()
	var entries []fs.FileInfo
	if ok {
		// Cache hit.
		entries = v.([]fs.FileInfo)
	} else {
		// Cache miss.
		var err error
		start := time.Now()
		entries, err = lsTreeUncached(ctx, repo, commit, path, recurse)
		if err != nil {
			return nil, err
		}

		// It's only worthwhile to cache if the operation took a while and returned a lot of
		// data. This is a heuristic.
		if time.Since(start) > 500*time.Millisecond && len(entries) > 5000 {
			lsTreeRootCacheMu.Lock()
			lsTreeRootCache.Add(key, entries)
			lsTreeRootCacheMu.Unlock()
		}
	}
	return entries, nil
}

func lsTreeUncached(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
	if err := ensureAbsoluteCommit(commit); err != nil {
		return nil, err
	}

	// Don't call filepath.Clean(path) because ReadDir needs to pass
	// path with a trailing slash.

	if err := checkSpecArgSafety(path); err != nil {
		return nil, err
	}

	args := []string{
		"ls-tree",
		"--long", // show size
		"--full-name",
		"-z",
		string(commit),
	}
	if recurse {
		args = append(args, "-r", "-t")
	}
	if path != "" {
		args = append(args, "--", filepath.ToSlash(path))
	}
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contains(out, []byte("exists on disk, but not in")) {
			return nil, &os.PathError{Op: "ls-tree", Path: filepath.ToSlash(path), Err: os.ErrNotExist}
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}

	if len(out) == 0 {
		// If we are listing the empty root tree, we will have no output.
		if stdlibpath.Clean(path) == "." {
			return []fs.FileInfo{}, nil
		}
		return nil, &os.PathError{Op: "git ls-tree", Path: path, Err: os.ErrNotExist}
	}

	trimPath := strings.TrimPrefix(path, "./")
	lines := strings.Split(string(out), "\x00")
	fis := make([]fs.FileInfo, len(lines)-1)
	for i, line := range lines {
		if i == len(lines)-1 {
			// last entry is empty
			continue
		}

		tabPos := strings.IndexByte(line, '\t')
		if tabPos == -1 {
			return nil, errors.Errorf("invalid `git ls-tree` output: %q", out)
		}
		info := strings.SplitN(line[:tabPos], " ", 4)
		name := line[tabPos+1:]
		if len(name) < len(trimPath) {
			// This is in a submodule; return the original path to avoid a slice out of bounds panic
			// when setting the FileInfo._Name below.
			name = trimPath
		}

		if len(info) != 4 {
			return nil, errors.Errorf("invalid `git ls-tree` output: %q", out)
		}
		typ := info[1]
		sha := info[2]
		if !IsAbsoluteRevision(sha) {
			return nil, errors.Errorf("invalid `git ls-tree` SHA output: %q", sha)
		}
		oid, err := decodeOID(sha)
		if err != nil {
			return nil, err
		}

		sizeStr := strings.TrimSpace(info[3])
		var size int64
		if sizeStr != "-" {
			// Size of "-" indicates a dir or submodule.
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size < 0 {
				return nil, errors.Errorf("invalid `git ls-tree` size output: %q (error: %s)", sizeStr, err)
			}
		}

		var sys interface{}
		modeVal, err := strconv.ParseInt(info[0], 8, 32)
		if err != nil {
			return nil, err
		}
		mode := os.FileMode(modeVal)
		switch typ {
		case "blob":
			const gitModeSymlink = 020000
			if mode&gitModeSymlink != 0 {
				mode = os.ModeSymlink
			} else {
				// Regular file.
				mode = mode | 0644
			}
		case "commit":
			mode = mode | ModeSubmodule
			cmd := gitserver.DefaultClient.Command("git", "show", fmt.Sprintf("%s:.gitmodules", commit))
			cmd.Repo = repo
			var submodule Submodule
			if out, err := cmd.Output(ctx); err == nil {

				var cfg config.Config
				err := config.NewDecoder(bytes.NewBuffer(out)).Decode(&cfg)
				if err != nil {
					return nil, errors.Errorf("error parsing .gitmodules: %s", err)
				}

				submodule.Path = cfg.Section("submodule").Subsection(name).Option("path")
				submodule.URL = cfg.Section("submodule").Subsection(name).Option("url")
			}
			submodule.CommitID = api.CommitID(oid.String())
			sys = submodule
		case "tree":
			mode = mode | os.ModeDir
		}

		if sys == nil {
			// Some callers might find it useful to know the object's OID.
			sys = objectInfo(oid)
		}

		fis[i] = &util.FileInfo{
			Name_: name, // full path relative to root (not just basename)
			Mode_: mode,
			Size_: size,
			Sys_:  sys,
		}
	}
	util.SortFileInfosByName(fis)

	return fis, nil
}

// ListFiles returns a list of root-relative file paths matching the given
// pattern in a particular commit of a repository.
func ListFiles(ctx context.Context, repo api.RepoName, commit api.CommitID, pattern *regexp.Regexp, checker authz.SubRepoPermissionChecker) (_ []string, err error) {
	cmd := gitserver.DefaultClient.Command("git", "ls-tree", "--name-only", "-r", string(commit), "--")
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
	if !checker.Enabled() {
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
func ListDirectoryChildren(ctx context.Context, repo api.RepoName, commit api.CommitID, dirnames []string) (map[string][]string, error) {
	args := []string{"ls-tree", "--name-only", string(commit), "--"}
	args = append(args, cleanDirectoriesForLsTree(dirnames)...)
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return parseDirectoryChildren(dirnames, strings.Split(string(out), "\n")), nil
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
