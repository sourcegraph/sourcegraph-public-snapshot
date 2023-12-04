package gitserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/mail"
	"os"
	stdlibpath "path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/golang/groupcache/lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DiffOptions struct {
	Repo api.RepoName

	// These fields must be valid <commit> inputs as defined by gitrevisions(7).
	Base string
	Head string

	// RangeType to be used for computing the diff: one of ".." or "..." (or unset: "").
	// For a nice visual explanation of ".." vs "...", see https://stackoverflow.com/a/46345364/2682729
	RangeType string

	Paths []string
}

// Diff returns an iterator that can be used to access the diff between two
// commits on a per-file basis. The iterator must be closed with Close when no
// longer required.
func (c *clientImplementor) Diff(ctx context.Context, opts DiffOptions) (_ *DiffFileIterator, err error) {
	ctx, _, endObservation := c.operations.diff.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			opts.Repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	// Rare case: the base is the empty tree, in which case we must use ..
	// instead of ... as the latter only works for commits.
	if opts.Base == DevNullSHA {
		opts.RangeType = ".."
	} else if opts.RangeType != ".." {
		opts.RangeType = "..."
	}

	rangeSpec := opts.Base + opts.RangeType + opts.Head
	if strings.HasPrefix(rangeSpec, "-") || strings.HasPrefix(rangeSpec, ".") {
		// We don't want to allow user input to add `git diff` command line
		// flags or refer to a file.
		return nil, errors.Errorf("invalid diff range argument: %q", rangeSpec)
	}
	args := append([]string{
		"diff",
		"--find-renames",
		// TODO(eseliger): Enable once we have support for copy detection in go-diff
		// and actually expose a `isCopy` field in the api, otherwise this
		// information is thrown away anyways.
		// "--find-copies",
		"--full-index",
		"--inter-hunk-context=3",
		"--no-prefix",
		rangeSpec,
		"--",
	}, opts.Paths...)

	rdr, err := c.gitCommand(opts.Repo, args...).StdoutReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "executing git diff")
	}

	return &DiffFileIterator{
		rdr:            rdr,
		mfdr:           diff.NewMultiFileDiffReader(rdr),
		fileFilterFunc: getFilterFunc(ctx, c.subRepoPermsChecker, opts.Repo),
	}, nil
}

type DiffFileIterator struct {
	rdr            io.ReadCloser
	mfdr           *diff.MultiFileDiffReader
	fileFilterFunc diffFileIteratorFilter
}

func NewDiffFileIterator(rdr io.ReadCloser) *DiffFileIterator {
	return &DiffFileIterator{
		rdr:  rdr,
		mfdr: diff.NewMultiFileDiffReader(rdr),
	}
}

type diffFileIteratorFilter func(fileName string) (bool, error)

func getFilterFunc(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName) diffFileIteratorFilter {
	if !authz.SubRepoEnabled(checker) {
		return nil
	}
	return func(fileName string) (bool, error) {
		shouldFilter, err := authz.FilterActorPath(ctx, checker, actor.FromContext(ctx), repo, fileName)
		if err != nil {
			return false, err
		}
		return shouldFilter, nil
	}
}

func (i *DiffFileIterator) Close() error {
	return i.rdr.Close()
}

// Next returns the next file diff. If no more diffs are available, the diff
// will be nil and the error will be io.EOF.
func (i *DiffFileIterator) Next() (*diff.FileDiff, error) {
	fd, err := i.mfdr.ReadFile()
	if err != nil {
		return fd, err
	}
	if i.fileFilterFunc != nil {
		if canRead, err := i.fileFilterFunc(fd.NewName); err != nil {
			return nil, err
		} else if !canRead {
			// go to next
			return i.Next()
		}
	}
	return fd, err
}

// ContributorOptions contains options for filtering contributor commit counts
type ContributorOptions struct {
	Range string // the range for which stats will be fetched
	After string // the date after which to collect commits
	Path  string // compute stats for commits that touch this path
}

func (o *ContributorOptions) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("range", o.Range),
		attribute.String("after", o.After),
		attribute.String("path", o.Path),
	}
}

func (c *clientImplementor) ContributorCount(ctx context.Context, repo api.RepoName, opt ContributorOptions) (_ []*gitdomain.ContributorCount, err error) {
	ctx, _, endObservation := c.operations.contributorCount.With(ctx, &err, observation.Args{Attrs: opt.Attrs(), MetricLabelValues: []string{c.scope}})
	defer endObservation(1, observation.Args{})

	if opt.Range == "" {
		opt.Range = "HEAD"
	}
	if err := checkSpecArgSafety(opt.Range); err != nil {
		return nil, err
	}

	// We split the individual args for the shortlog command instead of -sne for easier arg checking in the allowlist.
	args := []string{"shortlog", "-s", "-n", "-e", "--no-merges"}
	if opt.After != "" {
		args = append(args, "--after="+opt.After)
	}
	args = append(args, opt.Range, "--")
	if opt.Path != "" {
		args = append(args, opt.Path)
	}
	cmd := c.gitCommand(repo, args...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.Errorf("exec `git shortlog -s -n -e` failed: %v", err)
	}
	return parseShortLog(out)
}

// logEntryPattern is the regexp pattern that matches entries in the output of the `git shortlog
// -sne` command.
var logEntryPattern = lazyregexp.New(`^\s*([0-9]+)\s+(.*)$`)

func parseShortLog(out []byte) ([]*gitdomain.ContributorCount, error) {
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, nil
	}
	lines := bytes.Split(out, []byte{'\n'})
	results := make([]*gitdomain.ContributorCount, len(lines))
	for i, line := range lines {
		// example line: "1125\tJane Doe <jane@sourcegraph.com>"
		match := logEntryPattern.FindSubmatch(line)
		if match == nil {
			return nil, errors.Errorf("invalid git shortlog line: %q", line)
		}
		// example match: ["1125\tJane Doe <jane@sourcegraph.com>" "1125" "Jane Doe <jane@sourcegraph.com>"]
		count, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return nil, err
		}
		addr, err := lenientParseAddress(string(match[2]))
		if err != nil || addr == nil {
			addr = &mail.Address{Name: string(match[2])}
		}
		results[i] = &gitdomain.ContributorCount{
			Count: int32(count),
			Name:  addr.Name,
			Email: addr.Address,
		}
	}
	return results, nil
}

// lenientParseAddress is just like mail.ParseAddress, except that it treats
// the following somewhat-common malformed syntax where a user has misconfigured
// their email address as their name:
//
//	foo@gmail.com <foo@gmail.com>
//
// As a valid name, whereas mail.ParseAddress would return an error:
//
//	mail: expected single address, got "<foo@gmail.com>"
func lenientParseAddress(address string) (*mail.Address, error) {
	addr, err := mail.ParseAddress(address)
	if err != nil && strings.Contains(err.Error(), "expected single address") {
		p := strings.LastIndex(address, "<")
		if p == -1 {
			return addr, err
		}
		return &mail.Address{
			Name:    strings.TrimSpace(address[:p]),
			Address: strings.Trim(address[p:], " <>"),
		}, nil
	}
	return addr, err
}

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which
// could cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.Errorf("invalid git revision spec %q (begins with '-')", spec)
	}
	return nil
}

type CommitGraphOptions struct {
	Commit  string
	AllRefs bool
	Limit   int
	Since   *time.Time
} // please update LogFields if you add a field here

// CommitGraph returns the commit graph for the given repository as a mapping
// from a commit to its parents. If a commit is supplied, the returned graph will
// be rooted at the given commit. If a non-zero limit is supplied, at most that
// many commits will be returned.
func (c *clientImplementor) CommitGraph(ctx context.Context, repo api.RepoName, opts CommitGraphOptions) (_ *gitdomain.CommitGraph, err error) {
	ctx, _, endObservation := c.operations.commitGraph.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	args := []string{"log", "--pretty=%H %P", "--topo-order"}
	if opts.AllRefs {
		args = append(args, "--all")
	}
	if opts.Commit != "" {
		args = append(args, opts.Commit)
	}
	if opts.Since != nil {
		args = append(args, fmt.Sprintf("--since=%s", opts.Since.Format(time.RFC3339)))
	}
	if opts.Limit > 0 {
		args = append(args, fmt.Sprintf("-%d", opts.Limit))
	}

	cmd := c.gitCommand(repo, args...)

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return gitdomain.ParseCommitGraph(strings.Split(string(out), "\n")), nil
}

// CommitLog returns the repository commit log, including the file paths that were changed. The general approach to parsing
// is to separate the first line (the metadata line) from the remaining lines (the files), and then parse the metadata line
// into component parts separately.
func (c *clientImplementor) CommitLog(ctx context.Context, repo api.RepoName, after time.Time) (_ []CommitLog, err error) {
	ctx, _, endObservation := c.operations.commitLog.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.Stringer("after", after),
		},
	})
	defer endObservation(1, observation.Args{})

	args := []string{"log", "--pretty=format:%H<!>%ae<!>%an<!>%ad", "--name-only", "--topo-order", "--no-merges"}
	if !after.IsZero() {
		args = append(args, fmt.Sprintf("--after=%s", after.Format(time.RFC3339)))
	}

	cmd := c.gitCommand(repo, args...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "gitCommand %s", string(out))
	}

	var ls []CommitLog
	lines := strings.Split(string(out), "\n\n")

	for _, logOutput := range lines {
		partitions := strings.Split(logOutput, "\n")
		if len(partitions) < 2 {
			continue
		}
		metaLine := partitions[0]
		var changedFiles []string
		for _, pt := range partitions[1:] {
			if pt != "" {
				changedFiles = append(changedFiles, pt)
			}
		}

		parts := strings.Split(metaLine, "<!>")
		if len(parts) != 4 {
			continue
		}
		sha, authorEmail, authorName, timestamp := parts[0], parts[1], parts[2], parts[3]
		t, err := parseTimestamp(timestamp)
		if err != nil {
			return nil, errors.Wrapf(err, "parseTimestamp %s", timestamp)
		}
		ls = append(ls, CommitLog{
			SHA:          sha,
			AuthorEmail:  authorEmail,
			AuthorName:   authorName,
			Timestamp:    t,
			ChangedFiles: changedFiles,
		})
	}
	return ls, nil
}

func parseTimestamp(timestamp string) (time.Time, error) {
	layout := "Mon Jan 2 15:04:05 2006 -0700"
	t, err := time.Parse(layout, timestamp)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// DevNullSHA 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t
// tree /dev/null`, which is used as the base when computing the `git diff` of
// the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func (c *clientImplementor) DiffPath(ctx context.Context, repo api.RepoName, sourceCommit, targetCommit, path string) (_ []*diff.Hunk, err error) {
	ctx, _, endObservation := c.operations.diffPath.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	a := actor.FromContext(ctx)
	if hasAccess, err := authz.FilterActorPath(ctx, c.subRepoPermsChecker, a, repo, path); err != nil {
		return nil, err
	} else if !hasAccess {
		return nil, os.ErrNotExist
	}
	args := []string{"diff", sourceCommit, targetCommit, "--", path}
	reader, err := c.gitCommand(repo, args...).StdoutReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	output, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(output) == 0 {
		return nil, nil
	}

	d, err := diff.NewFileDiffReader(bytes.NewReader(output)).Read()
	if err != nil {
		return nil, err
	}
	return d.Hunks, nil
}

// DiffSymbols performs a diff command which is expected to be parsed by our symbols package
func (c *clientImplementor) DiffSymbols(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) (_ []byte, err error) {
	ctx, _, endObservation := c.operations.diffSymbols.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("commitA", string(commitA)),
			attribute.String("commitB", string(commitB)),
		},
	})
	defer endObservation(1, observation.Args{})

	// Ensure commits exist to provide a better error message
	_, err = c.ResolveRevisions(ctx, repo, []protocol.RevisionSpecifier{{RevSpec: string(commitA)}, {RevSpec: string(commitB)}})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lookup revisions for git diff on %s between %s and %s", repo, commitA, commitB)
	}

	command := c.gitCommand(repo, "diff", "-z", "--name-status", "--no-renames", string(commitA), string(commitB))
	out, err := command.Output(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to run git diff on %s between %s and %s", repo, commitA, commitB)
	}
	return out, nil
}

// ReadDir reads the contents of the named directory at commit.
func (c *clientImplementor) ReadDir(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) (_ []fs.FileInfo, err error) {
	ctx, _, endObservation := c.operations.readDir.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			commit.Attr(),
			attribute.String("path", path),
			attribute.Bool("recurse", recurse),
		},
	})
	defer endObservation(1, observation.Args{})

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	if path != "" {
		// Trailing slash is necessary to ls-tree under the dir (not just
		// to list the dir's tree entry in its parent dir).
		path = filepath.Clean(rel(path)) + "/"
	}
	files, err := c.lsTree(ctx, repo, commit, path, recurse)

	if err != nil || !authz.SubRepoEnabled(c.subRepoPermsChecker) {
		return files, err
	}

	a := actor.FromContext(ctx)
	filtered, filteringErr := authz.FilterActorFileInfos(ctx, c.subRepoPermsChecker, a, repo, files)
	if filteringErr != nil {
		return nil, errors.Wrap(err, "filtering paths")
	} else {
		return filtered, nil
	}
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
func (c *clientImplementor) lsTree(
	ctx context.Context,
	repo api.RepoName,
	commit api.CommitID,
	path string,
	recurse bool,
) (files []fs.FileInfo, err error) {
	if path != "" || !recurse {
		// Only cache the root recursive ls-tree.
		return c.lsTreeUncached(ctx, repo, commit, path, recurse)
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
		entries, err = c.lsTreeUncached(ctx, repo, commit, path, recurse)
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

type objectInfo gitdomain.OID

func (oid objectInfo) OID() gitdomain.OID { return gitdomain.OID(oid) }

// lStat returns a FileInfo describing the named file at commit. If the file is a
// symbolic link, the returned FileInfo describes the symbolic link. lStat makes
// no attempt to follow the link.
func (c *clientImplementor) lStat(ctx context.Context, repo api.RepoName, commit api.CommitID, path string) (_ fs.FileInfo, err error) {
	ctx, _, endObservation := c.operations.lstat.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			commit.Attr(),
			attribute.String("path", path),
		},
	})
	defer endObservation(1, observation.Args{})

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	path = filepath.Clean(rel(path))

	if path == "." {
		// Special case root, which is not returned by `git ls-tree`.
		obj, err := c.GetObject(ctx, repo, string(commit)+"^{tree}")
		if err != nil {
			return nil, err
		}
		return &fileutil.FileInfo{Mode_: os.ModeDir, Sys_: objectInfo(obj.ID)}, nil
	}

	fis, err := c.lsTree(ctx, repo, commit, path, false)
	if err != nil {
		return nil, err
	}
	if len(fis) == 0 {
		return nil, &os.PathError{Op: "ls-tree", Path: path, Err: os.ErrNotExist}
	}

	if !authz.SubRepoEnabled(c.subRepoPermsChecker) {
		return fis[0], nil
	}
	// Applying sub-repo permissions
	a := actor.FromContext(ctx)
	include, filteringErr := authz.FilterActorFileInfo(ctx, c.subRepoPermsChecker, a, repo, fis[0])
	if include && filteringErr == nil {
		return fis[0], nil
	} else {
		if filteringErr != nil {
			err = errors.Wrap(filteringErr, "filtering paths")
		} else {
			err = &os.PathError{Op: "ls-tree", Path: path, Err: os.ErrNotExist}
		}
		return nil, err
	}
}

func errorMessageTruncatedOutput(cmd []string, out []byte) string {
	const maxOutput = 5000

	message := fmt.Sprintf("git command %v failed", cmd)
	if len(out) > maxOutput {
		message += fmt.Sprintf(" (truncated output: %q, %d more)", out[:maxOutput], len(out)-maxOutput)
	} else {
		message += fmt.Sprintf(" (output: %q)", out)
	}

	return message
}

func (c *clientImplementor) lsTreeUncached(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
	if err := gitdomain.EnsureAbsoluteCommit(commit); err != nil {
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
	cmd := c.gitCommand(repo, args...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contains(out, []byte("exists on disk, but not in")) {
			return nil, &os.PathError{Op: "ls-tree", Path: filepath.ToSlash(path), Err: os.ErrNotExist}
		}

		message := errorMessageTruncatedOutput(cmd.Args(), out)
		return nil, errors.WithMessage(err, message)
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
		if !gitdomain.IsAbsoluteRevision(sha) {
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

		var sys any
		modeVal, err := strconv.ParseInt(info[0], 8, 32)
		if err != nil {
			return nil, err
		}
		mode := os.FileMode(modeVal)
		switch typ {
		case "blob":
			const gitModeSymlink = 0o20000
			if mode&gitModeSymlink != 0 {
				mode = os.ModeSymlink
			} else {
				// Regular file.
				mode = mode | 0o644
			}
		case "commit":
			mode = mode | gitdomain.ModeSubmodule
			cmd := c.gitCommand(repo, "show", fmt.Sprintf("%s:.gitmodules", commit))
			var submodule gitdomain.Submodule
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

		fis[i] = &fileutil.FileInfo{
			Name_: name, // full path relative to root (not just basename)
			Mode_: mode,
			Size_: size,
			Sys_:  sys,
		}
	}
	fileutil.SortFileInfosByName(fis)

	return fis, nil
}

func decodeOID(sha string) (gitdomain.OID, error) {
	oidBytes, err := hex.DecodeString(sha)
	if err != nil {
		return gitdomain.OID{}, err
	}
	var oid gitdomain.OID
	copy(oid[:], oidBytes)
	return oid, nil
}

func (c *clientImplementor) LogReverseEach(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdomain.LogEntry) error) (err error) {
	ctx, _, endObservation := c.operations.logReverseEach.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			api.RepoName(repo).Attr(),
			attribute.String("commit", commit),
		},
	})
	defer endObservation(1, observation.Args{})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	command := c.gitCommand(api.RepoName(repo), gitdomain.LogReverseArgs(n, commit)...)

	// We run a single `git log` command and stream the output while the repo is being processed, which
	// can take much longer than 1 minute (the default timeout).
	command.DisableTimeout()
	stdout, err := command.StdoutReader(ctx)
	if err != nil {
		return err
	}
	defer stdout.Close()

	return errors.Wrap(gitdomain.ParseLogReverseEach(stdout, onLogEntry), "ParseLogReverseEach")
}

// BlameOptions configures a blame.
type BlameOptions struct {
	NewestCommit     api.CommitID `json:",omitempty" url:",omitempty"`
	IgnoreWhitespace bool         `json:",omitempty" url:",omitempty"`
	StartLine        int          `json:",omitempty" url:",omitempty"` // 1-indexed start line (or 0 for beginning of file)
	EndLine          int          `json:",omitempty" url:",omitempty"` // 1-indexed end line (or 0 for end of file)
}

func (o *BlameOptions) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("newestCommit", string(o.NewestCommit)),
		attribute.Int("startLine", o.StartLine),
		attribute.Int("endLine", o.EndLine),
		attribute.Bool("ignoreWhitespace", o.IgnoreWhitespace),
	}
}

// A Hunk is a contiguous portion of a file associated with a commit.
type Hunk struct {
	StartLine int // 1-indexed start line number
	EndLine   int // 1-indexed end line number
	StartByte int // 0-indexed start byte position (inclusive)
	EndByte   int // 0-indexed end byte position (exclusive)
	api.CommitID
	Author   gitdomain.Signature
	Message  string
	Filename string
}

// StreamBlameFile returns Git blame information about a file.
func (c *clientImplementor) StreamBlameFile(ctx context.Context, repo api.RepoName, path string, opt *BlameOptions) (_ HunkReader, err error) {
	ctx, _, endObservation := c.operations.streamBlameFile.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: append([]attribute.KeyValue{
			repo.Attr(),
			attribute.String("path", path),
		}, opt.Attrs()...),
	})
	defer endObservation(1, observation.Args{})

	return streamBlameFileCmd(ctx, c.subRepoPermsChecker, repo, path, opt, c.gitserverGitCommandFunc(repo))
}

type errUnauthorizedStreamBlame struct {
	Repo api.RepoName
}

func (e errUnauthorizedStreamBlame) Unauthorized() bool {
	return true
}

func (e errUnauthorizedStreamBlame) Error() string {
	return fmt.Sprintf("not authorized (name=%s)", e.Repo)
}

func streamBlameFileCmd(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, path string, opt *BlameOptions, command gitCommandFunc) (HunkReader, error) {
	a := actor.FromContext(ctx)
	hasAccess, err := authz.FilterActorPath(ctx, checker, a, repo, path)
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, errUnauthorizedStreamBlame{Repo: repo}
	}
	if opt == nil {
		opt = &BlameOptions{}
	}
	if err := checkSpecArgSafety(string(opt.NewestCommit)); err != nil {
		return nil, err
	}

	args := []string{"blame", "--porcelain", "--incremental"}
	if opt.IgnoreWhitespace {
		args = append(args, "-w")
	}
	if opt.StartLine != 0 || opt.EndLine != 0 {
		args = append(args, fmt.Sprintf("-L%d,%d", opt.StartLine, opt.EndLine))
	}
	args = append(args, string(opt.NewestCommit), "--", filepath.ToSlash(path))

	rc, err := command(args).StdoutReader(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed", args))
	}

	return newBlameHunkReader(rc), nil
}

// BlameFile returns Git blame information about a file.
func (c *clientImplementor) BlameFile(ctx context.Context, repo api.RepoName, path string, opt *BlameOptions) (_ []*Hunk, err error) {
	ctx, _, endObservation := c.operations.blameFile.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: append([]attribute.KeyValue{
			repo.Attr(),
			attribute.String("path", path),
		}, opt.Attrs()...),
	})
	defer endObservation(1, observation.Args{})

	return blameFileCmd(ctx, c.subRepoPermsChecker, c.gitserverGitCommandFunc(repo), path, opt, repo)
}

func blameFileCmd(ctx context.Context, checker authz.SubRepoPermissionChecker, command gitCommandFunc, path string, opt *BlameOptions, repo api.RepoName) ([]*Hunk, error) {
	a := actor.FromContext(ctx)
	if hasAccess, err := authz.FilterActorPath(ctx, checker, a, repo, path); err != nil || !hasAccess {
		return nil, err
	}
	if opt == nil {
		opt = &BlameOptions{}
	}
	if err := checkSpecArgSafety(string(opt.NewestCommit)); err != nil {
		return nil, err
	}

	args := []string{"blame", "--porcelain"}
	if opt.IgnoreWhitespace {
		args = append(args, "-w")
	}
	if opt.StartLine != 0 || opt.EndLine != 0 {
		args = append(args, fmt.Sprintf("-L%d,%d", opt.StartLine, opt.EndLine))
	}
	args = append(args, string(opt.NewestCommit), "--", filepath.ToSlash(path))

	out, err := command(args).Output(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", args, out))
	}
	if len(out) == 0 {
		return nil, nil
	}

	return parseGitBlameOutput(string(out))
}

// parseGitBlameOutput parses the output of `git blame -w --porcelain`
func parseGitBlameOutput(out string) ([]*Hunk, error) {
	commits := make(map[string]gitdomain.Commit)
	filenames := make(map[string]string)
	hunks := make([]*Hunk, 0)
	remainingLines := strings.Split(out[:len(out)-1], "\n")
	byteOffset := 0
	for len(remainingLines) > 0 {
		// Consume hunk
		hunkHeader := strings.Split(remainingLines[0], " ")
		if len(hunkHeader) != 4 {
			return nil, errors.Errorf("Expected at least 4 parts to hunkHeader, but got: '%s'", hunkHeader)
		}
		commitID := hunkHeader[0]
		lineNoCur, _ := strconv.Atoi(hunkHeader[2])
		nLines, _ := strconv.Atoi(hunkHeader[3])
		hunk := &Hunk{
			CommitID:  api.CommitID(commitID),
			StartLine: lineNoCur,
			EndLine:   lineNoCur + nLines,
			StartByte: byteOffset,
		}

		if _, in := commits[commitID]; in {
			// Already seen commit
			byteOffset += len(remainingLines[1])
			remainingLines = remainingLines[2:]
		} else {
			// New commit
			author := strings.Join(strings.Split(remainingLines[1], " ")[1:], " ")
			email := strings.Join(strings.Split(remainingLines[2], " ")[1:], " ")
			if len(email) >= 2 && email[0] == '<' && email[len(email)-1] == '>' {
				email = email[1 : len(email)-1]
			}
			authorTime, err := strconv.ParseInt(strings.Join(strings.Split(remainingLines[3], " ")[1:], " "), 10, 64)
			if err != nil {
				return nil, errors.Errorf("Failed to parse author-time %q", remainingLines[3])
			}
			summary := strings.Join(strings.Split(remainingLines[9], " ")[1:], " ")
			commit := gitdomain.Commit{
				ID:      api.CommitID(commitID),
				Message: gitdomain.Message(summary),
				Author: gitdomain.Signature{
					Name:  author,
					Email: email,
					Date:  time.Unix(authorTime, 0).UTC(),
				},
			}

			for i := 10; i < 13 && i < len(remainingLines); i++ {
				if strings.HasPrefix(remainingLines[i], "filename ") {
					filenames[commitID] = strings.SplitN(remainingLines[i], " ", 2)[1]
					break
				}
			}

			if len(remainingLines) >= 13 && strings.HasPrefix(remainingLines[10], "previous ") {
				byteOffset += len(remainingLines[12])
				remainingLines = remainingLines[13:]
			} else if len(remainingLines) >= 13 && remainingLines[10] == "boundary" {
				byteOffset += len(remainingLines[12])
				remainingLines = remainingLines[13:]
			} else if len(remainingLines) >= 12 {
				byteOffset += len(remainingLines[11])
				remainingLines = remainingLines[12:]
			} else if len(remainingLines) == 11 {
				// Empty file
				remainingLines = remainingLines[11:]
			} else {
				return nil, errors.Errorf("Unexpected number of remaining lines (%d):\n%s", len(remainingLines), "  "+strings.Join(remainingLines, "\n  "))
			}

			commits[commitID] = commit
		}

		if commit, present := commits[commitID]; present {
			// Should always be present, but check just to avoid
			// panicking in case of a (somewhat likely) bug in our
			// git-blame parser above.
			hunk.CommitID = commit.ID
			hunk.Author = commit.Author
			hunk.Message = string(commit.Message)
		}

		if filename, present := filenames[commitID]; present {
			hunk.Filename = filename
		}

		// Consume remaining lines in hunk
		for i := 1; i < nLines; i++ {
			byteOffset += len(remainingLines[1])
			remainingLines = remainingLines[2:]
		}

		hunk.EndByte = byteOffset
		hunks = append(hunks, hunk)
	}

	return hunks, nil
}

func (c *clientImplementor) gitserverGitCommandFunc(repo api.RepoName) gitCommandFunc {
	return func(args []string) GitCommand {
		return c.gitCommand(repo, args...)
	}
}

// gitCommandFunc is a func that creates a new executable Git command.
type gitCommandFunc func(args []string) GitCommand

// ResolveRevisionOptions configure how we resolve revisions.
// The zero value should contain appropriate default values.
type ResolveRevisionOptions struct {
	NoEnsureRevision bool // do not try to fetch from remote if revision doesn't exist locally
}

var resolveRevisionCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_resolve_revision_total",
	Help: "The number of times we call internal/vcs/git/ResolveRevision",
}, []string{"ensure_revision"})

// ResolveRevision will return the absolute commit for a commit-ish spec. If spec is empty, HEAD is
// used.
//
// Error cases:
// * Repo does not exist: gitdomain.RepoNotExistError
// * Commit does not exist: gitdomain.RevisionNotFoundError
// * Empty repository: gitdomain.RevisionNotFoundError
// * Other unexpected errors.
func (c *clientImplementor) ResolveRevision(ctx context.Context, repo api.RepoName, spec string, opt ResolveRevisionOptions) (_ api.CommitID, err error) {
	labelEnsureRevisionValue := "true"
	if opt.NoEnsureRevision {
		labelEnsureRevisionValue = "false"
	}
	resolveRevisionCounter.WithLabelValues(labelEnsureRevisionValue).Inc()

	ctx, _, endObservation := c.operations.resolveRevision.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("spec", spec),
			attribute.Bool("noEnsureRevision", opt.NoEnsureRevision),
		},
	})
	defer endObservation(1, observation.Args{})

	if err := checkSpecArgSafety(spec); err != nil {
		return "", err
	}
	if spec == "" {
		spec = "HEAD"
	}
	if spec != "HEAD" {
		// "git rev-parse HEAD^0" is slower than "git rev-parse HEAD"
		// since it checks that the resolved git object exists. We can
		// assume it exists for HEAD, but for other commits we should
		// check.
		spec = spec + "^0"
	}

	cmd := c.gitCommand(repo, "rev-parse", spec)
	cmd.SetEnsureRevision(spec)

	// We don't ever need to ensure that HEAD is in git-server.
	// HEAD is always there once a repo is cloned
	// (except empty repos, but we don't need to ensure revision on those).
	if opt.NoEnsureRevision || spec == "HEAD" {
		cmd.SetEnsureRevision("")
	}

	return runRevParse(ctx, cmd, spec)
}

// runRevParse sends the git rev-parse command to gitserver. It interprets
// missing revision responses and converts them into RevisionNotFoundError.
func runRevParse(ctx context.Context, cmd GitCommand, spec string) (api.CommitID, error) {
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		if gitdomain.IsRepoNotExist(err) {
			return "", err
		}
		if bytes.Contains(stderr, []byte("unknown revision")) {
			return "", &gitdomain.RevisionNotFoundError{Repo: cmd.Repo(), Spec: spec}
		}
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (stderr: %q)", cmd.Args(), stderr))
	}
	commit := api.CommitID(bytes.TrimSpace(stdout))
	if !gitdomain.IsAbsoluteRevision(string(commit)) {
		if commit == "HEAD" {
			// We don't verify the existence of HEAD (see above comments), but
			// if HEAD doesn't point to anything git just returns `HEAD` as the
			// output of rev-parse. An example where this occurs is an empty
			// repository.
			return "", &gitdomain.RevisionNotFoundError{Repo: cmd.Repo(), Spec: spec}
		}
		return "", &gitdomain.BadCommitError{Spec: spec, Commit: commit, Repo: cmd.Repo()}
	}
	return commit, nil
}

// LsFiles returns the output of `git ls-files`.
func (c *clientImplementor) LsFiles(ctx context.Context, repo api.RepoName, commit api.CommitID, pathspecs ...gitdomain.Pathspec) (_ []string, err error) {
	ctx, _, endObservation := c.operations.lsFiles.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("commit", string(commit)),
			attribute.Bool("hasPathSpecs", len(pathspecs) > 0),
		},
	})
	defer endObservation(1, observation.Args{})

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

	cmd := c.gitCommand(repo, args...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}

	files := strings.Split(string(out), "\x00")
	// Drop trailing empty string
	if len(files) > 0 && files[len(files)-1] == "" {
		files = files[:len(files)-1]
	}
	return filterPaths(ctx, c.subRepoPermsChecker, repo, files)
}

// 🚨 SECURITY: All git methods that deal with file or path access need to have
// sub-repo permissions applied
func filterPaths(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, paths []string) ([]string, error) {
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
func (c *clientImplementor) ListDirectoryChildren(
	ctx context.Context,
	repo api.RepoName,
	commit api.CommitID,
	dirnames []string,
) (_ map[string][]string, err error) {
	ctx, _, endObservation := c.operations.listDirectoryChildren.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("commit", string(commit)),
			attribute.Int("dirs", len(dirnames)),
		},
	})
	defer endObservation(1, observation.Args{})

	args := []string{"ls-tree", "--name-only", string(commit), "--"}
	args = append(args, cleanDirectoriesForLsTree(dirnames)...)
	cmd := c.gitCommand(repo, args...)

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	paths := strings.Split(string(out), "\n")
	if authz.SubRepoEnabled(c.subRepoPermsChecker) {
		paths, err = authz.FilterActorPaths(ctx, c.subRepoPermsChecker, actor.FromContext(ctx), repo, paths)
		if err != nil {
			return nil, err
		}
	}
	return parseDirectoryChildren(dirnames, paths), nil
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

// ListTags returns a list of all tags in the repository. If commitObjs is non-empty, only all tags pointing at those commits are returned.
func (c *clientImplementor) ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error) {
	ctx, _, endObservation := c.operations.listTags.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.StringSlice("commitObjs", commitObjs),
		},
	})
	defer endObservation(1, observation.Args{})

	// Support both lightweight tags and tag objects. For creatordate, use an %(if) to prefer the
	// taggerdate for tag objects, otherwise use the commit's committerdate (instead of just always
	// using committerdate).
	args := []string{"tag", "--list", "--sort", "-creatordate", "--format", "%(if)%(*objectname)%(then)%(*objectname)%(else)%(objectname)%(end)%00%(refname:short)%00%(if)%(creatordate:unix)%(then)%(creatordate:unix)%(else)%(*creatordate:unix)%(end)"}

	for _, commit := range commitObjs {
		args = append(args, "--points-at", commit)
	}

	cmd := c.gitCommand(repo, args...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if gitdomain.IsRepoNotExist(err) {
			return nil, err
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}

	return parseTags(out)
}

func parseTags(in []byte) ([]*gitdomain.Tag, error) {
	in = bytes.TrimSuffix(in, []byte("\n")) // remove trailing newline
	if len(in) == 0 {
		return nil, nil // no tags
	}
	lines := bytes.Split(in, []byte("\n"))
	tags := make([]*gitdomain.Tag, len(lines))
	for i, line := range lines {
		parts := bytes.SplitN(line, []byte("\x00"), 3)
		if len(parts) != 3 {
			return nil, errors.Errorf("invalid git tag list output line: %q", line)
		}

		tag := &gitdomain.Tag{
			Name:     string(parts[1]),
			CommitID: api.CommitID(parts[0]),
		}

		date, err := strconv.ParseInt(string(parts[2]), 10, 64)
		if err == nil {
			tag.CreatorDate = time.Unix(date, 0).UTC()
		}

		tags[i] = tag
	}
	return tags, nil
}

// GetDefaultBranch returns the name of the default branch and the commit it's
// currently at from the given repository. If short is true, then `main` instead
// of `refs/heads/main` would be returned.
//
// If the repository is empty or currently being cloned, empty values and no
// error are returned.
func (c *clientImplementor) GetDefaultBranch(ctx context.Context, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error) {
	ctx, _, endObservation := c.operations.getDefaultBranch.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	args := []string{"symbolic-ref", "HEAD"}
	if short {
		args = append(args, "--short")
	}
	cmd := c.gitCommand(repo, args...)
	refBytes, _, err := cmd.DividedOutput(ctx)
	exitCode := cmd.ExitStatus()
	if exitCode != 0 && err != nil {
		err = nil // the error must just indicate that the exit code was nonzero
	}
	refName = string(bytes.TrimSpace(refBytes))

	if err == nil && exitCode == 0 {
		// Check that our repo is not empty
		commit, err = c.ResolveRevision(ctx, repo, "HEAD", ResolveRevisionOptions{NoEnsureRevision: true})
	}

	// If we fail to get the default branch due to cloning or being empty, we return nothing.
	if err != nil {
		if gitdomain.IsCloneInProgress(err) || errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			return "", "", nil
		}
		return "", "", err
	}

	return refName, commit, nil
}

// MergeBase returns the merge base commit for the specified commits.
func (c *clientImplementor) MergeBase(ctx context.Context, repo api.RepoName, a, b api.CommitID) (_ api.CommitID, err error) {
	ctx, _, endObservation := c.operations.mergeBase.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.String("a", string(a)),
			attribute.String("b", string(b)),
		},
	})
	defer endObservation(1, observation.Args{})

	cmd := c.gitCommand(repo, "merge-base", "--", string(a), string(b))
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}
	return api.CommitID(bytes.TrimSpace(out)), nil
}

// RevList makes a git rev-list call and iterates through the resulting commits, calling the provided onCommit function for each.
func (c *clientImplementor) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) (err error) {
	ctx, _, endObservation := c.operations.revList.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.String("repo", repo),
			attribute.String("commit", commit),
		},
	})
	defer endObservation(1, observation.Args{})

	command := c.gitCommand(api.RepoName(repo), RevListArgs(commit)...)
	command.DisableTimeout()
	stdout, err := command.StdoutReader(ctx)
	if err != nil {
		return err
	}
	defer stdout.Close()

	return gitdomain.RevListEach(stdout, onCommit)
}

func RevListArgs(givenCommit string) []string {
	return []string{"rev-list", "--first-parent", givenCommit}
}

// GetBehindAhead returns the behind/ahead commit counts information for right vs. left (both Git
// revspecs).
func (c *clientImplementor) GetBehindAhead(ctx context.Context, repo api.RepoName, left, right string) (_ *gitdomain.BehindAhead, err error) {
	ctx, _, endObservation := c.operations.getBehindAhead.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("left", left),
			attribute.String("right", right),
		},
	})
	defer endObservation(1, observation.Args{})

	if err := checkSpecArgSafety(left); err != nil {
		return nil, err
	}
	if err := checkSpecArgSafety(right); err != nil {
		return nil, err
	}

	cmd := c.gitCommand(repo, "rev-list", "--count", "--left-right", fmt.Sprintf("%s...%s", left, right))
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, err
	}
	behindAhead := strings.Split(strings.TrimSuffix(string(out), "\n"), "\t")
	b, err := strconv.ParseUint(behindAhead[0], 10, 0)
	if err != nil {
		return nil, err
	}
	a, err := strconv.ParseUint(behindAhead[1], 10, 0)
	if err != nil {
		return nil, err
	}
	return &gitdomain.BehindAhead{Behind: uint32(b), Ahead: uint32(a)}, nil
}

// ReadFile returns the first maxBytes of the named file at commit. If maxBytes <= 0, the entire
// file is read. (If you just need to check a file's existence, use Stat, not ReadFile.)
func (c *clientImplementor) ReadFile(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) (_ []byte, err error) {
	ctx, _, endObservation := c.operations.readFile.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			commit.Attr(),
			attribute.String("name", name),
		},
	})
	defer endObservation(1, observation.Args{})

	br, err := c.NewFileReader(ctx, repo, commit, name)
	if err != nil {
		return nil, err
	}
	defer br.Close()

	r := io.Reader(br)
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// NewFileReader returns an io.ReadCloser reading from the named file at commit.
// The caller should always close the reader after use
func (c *clientImplementor) NewFileReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) (_ io.ReadCloser, err error) {
	// TODO: this does not capture the lifetime of the request since we return a reader
	ctx, _, endObservation := c.operations.newFileReader.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			commit.Attr(),
			attribute.String("name", name),
		},
	})
	defer endObservation(1, observation.Args{})

	a := actor.FromContext(ctx)
	if hasAccess, err := authz.FilterActorPath(ctx, c.subRepoPermsChecker, a, repo, name); err != nil {
		return nil, err
	} else if !hasAccess {
		return nil, os.ErrNotExist
	}

	name = rel(name)
	br, err := c.newBlobReader(ctx, repo, commit, name)
	if err != nil {
		return nil, errors.Wrapf(err, "getting blobReader for %q", name)
	}
	return br, nil
}

// blobReader, which should be created using newBlobReader, is a struct that allows
// us to get a ReadCloser to a specific named file at a specific commit
type blobReader struct {
	c      *clientImplementor
	ctx    context.Context
	repo   api.RepoName
	commit api.CommitID
	name   string
	cmd    GitCommand
	rc     io.ReadCloser
}

func (c *clientImplementor) blobOID(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) (string, error) {
	// Note: when our git is new enough we can just use --object-only
	out, err := c.gitCommand(repo, "ls-tree", string(commit), "--", name).Output(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to lookup blob OID")
	}

	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return "", &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
	}

	// 100644 blob 3bad331187e39c05c78a9b5e443689f78f4365a7	README.md
	fields := bytes.Fields(out)
	if len(fields) < 3 {
		return "", errors.Newf("unexpected output while parsing blob OID: %q", string(out))
	}
	return string(fields[2]), nil
}

func (c *clientImplementor) newBlobReader(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) (*blobReader, error) {
	if err := gitdomain.EnsureAbsoluteCommit(commit); err != nil {
		return nil, err
	}

	var cmd GitCommand
	if strings.Contains(name, "..") {
		// We special case ".." in path to running a less efficient two
		// commands. For other paths we can rely on the faster git show.
		//
		// git show will try and resolve revisions on anything containing
		// "..". Depending on what branches/files exist, this can lead to:
		//
		//   - error: object $SHA is a tree, not a commit
		//   - fatal: Invalid symmetric difference expression $SHA:$name
		//   - outputting a diff instead of the file
		//
		// The last point is a security issue for repositories with sub-repo
		// permissions since the diff will not be filtered.
		blobOID, err := c.blobOID(ctx, repo, commit, name)
		if err != nil {
			return nil, err
		}
		cmd = c.gitCommand(repo, "cat-file", "-p", blobOID)
	} else {
		// Otherwise we can rely on a single command git show sha:name.
		cmd = c.gitCommand(repo, "show", string(commit)+":"+name)
	}

	stdout, err := cmd.StdoutReader(ctx)
	if err != nil {
		return nil, err
	}

	return &blobReader{
		c:      c,
		ctx:    ctx,
		repo:   repo,
		commit: commit,
		name:   name,
		cmd:    cmd,
		rc:     stdout,
	}, nil
}

func (br *blobReader) Read(p []byte) (int, error) {
	n, err := br.rc.Read(p)
	if err != nil {
		return n, br.convertError(err)
	}
	return n, nil
}

func (br *blobReader) Close() error {
	return br.rc.Close()
}

// convertError converts an error returned from 'git show' into a more appropriate error type
func (br *blobReader) convertError(err error) error {
	if err == nil {
		return nil
	}
	if err == io.EOF {
		return err
	}
	if strings.Contains(err.Error(), "exists on disk, but not in") || strings.Contains(err.Error(), "does not exist") {
		return &os.PathError{Op: "open", Path: br.name, Err: os.ErrNotExist}
	}
	if strings.Contains(err.Error(), "fatal: bad object ") {
		// Could be a git submodule.
		fi, err := br.c.Stat(br.ctx, br.repo, br.commit, br.name)
		if err != nil {
			return err
		}
		// Return EOF for a submodule for now which indicates zero content
		if fi.Mode()&gitdomain.ModeSubmodule != 0 {
			return io.EOF
		}
	}
	return errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", br.cmd.Args(), err))
}

// Stat returns a FileInfo describing the named file at commit.
func (c *clientImplementor) Stat(ctx context.Context, repo api.RepoName, commit api.CommitID, path string) (_ fs.FileInfo, err error) {
	ctx, _, endObservation := c.operations.stat.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			commit.Attr(),
			attribute.String("path", path),
		},
	})
	defer endObservation(1, observation.Args{})

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	path = rel(path)

	fi, err := c.lStat(ctx, repo, commit, path)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

// CommitsOptions specifies options for Commits.
type CommitsOptions struct {
	Range string // commit range (revspec, "A..B", "A...B", etc.)

	N    uint // limit the number of returned commits to this many (0 means no limit)
	Skip uint // skip this many commits at the beginning

	MessageQuery string // include only commits whose commit message contains this substring

	Author string // include only commits whose author matches this
	After  string // include only commits after this date
	Before string // include only commits before this date

	Reverse   bool // Whether or not commits should be given in reverse order (optional)
	DateOrder bool // Whether or not commits should be sorted by date (optional)

	Path string // only commits modifying the given path are selected (optional)

	Follow bool // follow the history of the path beyond renames (works only for a single path)

	// When true we opt out of attempting to fetch missing revisions
	NoEnsureRevision bool

	// When true return the names of the files changed in the commit
	NameOnly bool
}

var recordGetCommitQueries = os.Getenv("RECORD_GET_COMMIT_QUERIES") == "1"

// getCommit returns the commit with the given id.
func (c *clientImplementor) getCommit(ctx context.Context, repo api.RepoName, id api.CommitID, opt ResolveRevisionOptions) (_ *gitdomain.Commit, err error) {
	if honey.Enabled() && recordGetCommitQueries {
		defer func() {
			ev := honey.NewEvent("getCommit")
			ev.SetSampleRate(10) // 1 in 10
			ev.AddField("repo", repo)
			ev.AddField("commit", id)
			ev.AddField("no_ensure_revision", opt.NoEnsureRevision)
			ev.AddField("actor", actor.FromContext(ctx).UIDString())

			q, _ := ctx.Value(trace.GraphQLQueryKey).(string)
			ev.AddField("query", q)

			if err != nil {
				ev.AddField("error", err.Error())
			}

			_ = ev.Send()
		}()
	}

	if err := checkSpecArgSafety(string(id)); err != nil {
		return nil, err
	}

	commitOptions := CommitsOptions{
		Range:            string(id),
		N:                1,
		NoEnsureRevision: opt.NoEnsureRevision,
	}
	commitOptions = addNameOnly(commitOptions, c.subRepoPermsChecker)

	commits, err := c.commitLog(ctx, repo, commitOptions)
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, &gitdomain.RevisionNotFoundError{Repo: repo, Spec: string(id)}
	}
	if len(commits) != 1 {
		return nil, errors.Errorf("git log: expected 1 commit, got %d", len(commits))
	}

	return commits[0], nil
}

// GetCommit returns the commit with the given commit ID, or ErrCommitNotFound if no such commit
// exists.
//
// The remoteURLFunc is called to get the Git remote URL if it's not set in repo and if it is
// needed. The Git remote URL is only required if the gitserver doesn't already contain a clone of
// the repository or if the commit must be fetched from the remote.
func (c *clientImplementor) GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID, opt ResolveRevisionOptions) (_ *gitdomain.Commit, err error) {
	ctx, _, endObservation := c.operations.getCommit.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			id.Attr(),
			attribute.Bool("noEnsureRevision", opt.NoEnsureRevision),
		},
	})
	defer endObservation(1, observation.Args{})

	return c.getCommit(ctx, repo, id, opt)
}

// Commits returns all commits matching the options.
func (c *clientImplementor) Commits(ctx context.Context, repo api.RepoName, opt CommitsOptions) (_ []*gitdomain.Commit, err error) {
	opt = addNameOnly(opt, c.subRepoPermsChecker)
	ctx, _, endObservation := c.operations.commits.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("opts", fmt.Sprintf("%#v", opt)),
		},
	})
	defer endObservation(1, observation.Args{})

	if err := checkSpecArgSafety(opt.Range); err != nil {
		return nil, err
	}
	return c.commitLog(ctx, repo, opt)
}

func filterCommits(ctx context.Context, checker authz.SubRepoPermissionChecker, commits []*wrappedCommit, repoName api.RepoName) ([]*gitdomain.Commit, error) {
	if !authz.SubRepoEnabled(checker) {
		return unWrapCommits(commits), nil
	}
	filtered := make([]*gitdomain.Commit, 0, len(commits))
	for _, commit := range commits {
		if hasAccess, err := hasAccessToCommit(ctx, commit, repoName, checker); hasAccess {
			filtered = append(filtered, commit.Commit)
		} else if err != nil {
			return nil, err
		}
	}
	return filtered, nil
}

func unWrapCommits(wrappedCommits []*wrappedCommit) []*gitdomain.Commit {
	commits := make([]*gitdomain.Commit, 0, len(wrappedCommits))
	for _, wc := range wrappedCommits {
		commits = append(commits, wc.Commit)
	}
	return commits
}

func hasAccessToCommit(ctx context.Context, commit *wrappedCommit, repoName api.RepoName, checker authz.SubRepoPermissionChecker) (bool, error) {
	a := actor.FromContext(ctx)
	if commit.files == nil || len(commit.files) == 0 {
		return true, nil // If commit has no files, assume user has access to view the commit.
	}
	for _, fileName := range commit.files {
		if hasAccess, err := authz.FilterActorPath(ctx, checker, a, repoName, fileName); err != nil {
			return false, err
		} else if hasAccess {
			// if the user has access to one file modified in the commit, they have access to view the commit
			return true, nil
		}
	}
	return false, nil
}

// CommitsUniqueToBranch returns a map from commits that exist on a particular
// branch in the given repository to their committer date. This set of commits is
// determined by listing `{branchName} ^HEAD`, which is interpreted as: all
// commits on {branchName} not also on the tip of the default branch. If the
// supplied branch name is the default branch, then this method instead returns
// all commits reachable from HEAD.
func (c *clientImplementor) CommitsUniqueToBranch(ctx context.Context, repo api.RepoName, branchName string, isDefaultBranch bool, maxAge *time.Time) (_ map[string]time.Time, err error) {
	ctx, _, endObservation := c.operations.commitsUniqueToBranch.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("branch", branchName),
		},
	})
	defer endObservation(1, observation.Args{})

	args := []string{"log", "--pretty=format:%H:%cI"}
	if maxAge != nil {
		args = append(args, fmt.Sprintf("--after=%s", *maxAge))
	}
	if isDefaultBranch {
		args = append(args, "HEAD")
	} else {
		args = append(args, branchName, "^HEAD")
	}

	cmd := c.gitCommand(repo, args...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	commits, err := parseCommitsUniqueToBranch(strings.Split(string(out), "\n"))
	if authz.SubRepoEnabled(c.subRepoPermsChecker) && err == nil {
		return c.filterCommitsUniqueToBranch(ctx, repo, commits), nil
	}
	return commits, err
}

func (c *clientImplementor) filterCommitsUniqueToBranch(ctx context.Context, repo api.RepoName, commitsMap map[string]time.Time) map[string]time.Time {
	filtered := make(map[string]time.Time, len(commitsMap))
	for commitID, timeStamp := range commitsMap {
		_, err := c.GetCommit(ctx, repo, api.CommitID(commitID), ResolveRevisionOptions{
			// The commits are returned from git log, so we are sure they exist and
			// don't need to ensure the revision.
			NoEnsureRevision: true,
		})
		if !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			filtered[commitID] = timeStamp
		}
	}
	return filtered
}

func parseCommitsUniqueToBranch(lines []string) (_ map[string]time.Time, err error) {
	commitDates := make(map[string]time.Time, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, errors.Errorf(`unexpected output from git log "%s"`, line)
		}

		duration, err := time.Parse(time.RFC3339, parts[1])
		if err != nil {
			return nil, errors.Errorf(`unexpected output from git log (bad date format) "%s"`, line)
		}

		commitDates[parts[0]] = duration
	}

	return commitDates, nil
}

// HasCommitAfter indicates the staleness of a repository. It returns a boolean indicating if a repository
// contains a commit past a specified date.
func (c *clientImplementor) HasCommitAfter(ctx context.Context, repo api.RepoName, date string, revspec string) (_ bool, err error) {
	ctx, _, endObservation := c.operations.hasCommitAfter.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("date", date),
			attribute.String("revSpec", revspec),
		},
	})
	defer endObservation(1, observation.Args{})

	if authz.SubRepoEnabled(c.subRepoPermsChecker) {
		return c.hasCommitAfterWithFiltering(ctx, repo, date, revspec)
	}

	if revspec == "" {
		revspec = "HEAD"
	}

	commitid, err := c.ResolveRevision(ctx, repo, revspec, ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return false, err
	}

	args, err := commitLogArgs([]string{"rev-list", "--count"}, CommitsOptions{
		N:     1,
		After: date,
		Range: string(commitid),
	})
	if err != nil {
		return false, err
	}

	cmd := c.gitCommand(repo, args...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return false, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}

	out = bytes.TrimSpace(out)
	n, err := strconv.Atoi(string(out))
	return n > 0, err
}

func (c *clientImplementor) hasCommitAfterWithFiltering(ctx context.Context, repo api.RepoName, date, revspec string) (bool, error) {
	if commits, err := c.Commits(ctx, repo, CommitsOptions{After: date, Range: revspec}); err != nil {
		return false, err
	} else if len(commits) > 0 {
		return true, nil
	}
	return false, nil
}

func isBadObjectErr(output, obj string) bool {
	return output == "fatal: bad object "+obj
}

// commitLog returns a list of commits.
//
// The caller is responsible for doing checkSpecArgSafety on opt.Head and opt.Base.
func (c *clientImplementor) commitLog(ctx context.Context, repo api.RepoName, opt CommitsOptions) ([]*gitdomain.Commit, error) {
	wrappedCommits, err := c.getWrappedCommits(ctx, repo, opt)
	if err != nil {
		return nil, err
	}

	filtered, err := filterCommits(ctx, c.subRepoPermsChecker, wrappedCommits, repo)
	if err != nil {
		return nil, errors.Wrap(err, "filtering commits")
	}

	if needMoreCommits(filtered, wrappedCommits, opt, c.subRepoPermsChecker) {
		return c.getMoreCommits(ctx, repo, opt, filtered)
	}
	return filtered, err
}

func (c *clientImplementor) getWrappedCommits(ctx context.Context, repo api.RepoName, opt CommitsOptions) ([]*wrappedCommit, error) {
	args, err := commitLogArgs([]string{"log", logFormatWithoutRefs}, opt)
	if err != nil {
		return nil, err
	}

	cmd := c.gitCommand(repo, args...)
	if !opt.NoEnsureRevision {
		cmd.SetEnsureRevision(opt.Range)
	}
	wrappedCommits, err := runCommitLog(ctx, cmd, opt)
	if err != nil {
		return nil, err
	}
	return wrappedCommits, nil
}

func needMoreCommits(filtered []*gitdomain.Commit, commits []*wrappedCommit, opt CommitsOptions, checker authz.SubRepoPermissionChecker) bool {
	if !authz.SubRepoEnabled(checker) {
		return false
	}
	if opt.N == 0 || isRequestForSingleCommit(opt) {
		return false
	}
	if len(filtered) < len(commits) {
		return true
	}
	return false
}

func isRequestForSingleCommit(opt CommitsOptions) bool {
	return opt.Range != "" && opt.N == 1
}

// getMoreCommits handles the case where a specific number of commits was requested via CommitsOptions, but after sub-repo
// filtering, fewer than that requested number was left. This function requests the next N commits (where N was the number
// originally requested), filters the commits, and determines if this is at least N commits total after filtering. If not,
// the loop continues until N total filtered commits are collected _or_ there are no commits left to request.
func (c *clientImplementor) getMoreCommits(ctx context.Context, repo api.RepoName, opt CommitsOptions, baselineCommits []*gitdomain.Commit) ([]*gitdomain.Commit, error) {
	// We want to place an upper bound on the number of times we loop here so that we
	// don't hit pathological conditions where a lot of filtering has been applied.
	const maxIterations = 5

	totalCommits := make([]*gitdomain.Commit, 0, opt.N)
	for i := 0; i < maxIterations; i++ {
		if uint(len(totalCommits)) == opt.N {
			break
		}
		// Increment the Skip number to get the next N commits
		opt.Skip += opt.N
		wrappedCommits, err := c.getWrappedCommits(ctx, repo, opt)
		if err != nil {
			return nil, err
		}
		filtered, err := filterCommits(ctx, c.subRepoPermsChecker, wrappedCommits, repo)
		if err != nil {
			return nil, err
		}
		// join the new (filtered) commits with those already fetched (potentially truncating the list to have length N if necessary)
		totalCommits = joinCommits(baselineCommits, filtered, opt.N)
		baselineCommits = totalCommits
		if uint(len(wrappedCommits)) < opt.N {
			// No more commits available before filtering, so return current total commits (e.g. the last "page" of N commits has been reached)
			break
		}
	}
	return totalCommits, nil
}

func joinCommits(previous, next []*gitdomain.Commit, desiredTotal uint) []*gitdomain.Commit {
	allCommits := append(previous, next...)
	// ensure that we don't return more than what was requested
	if uint(len(allCommits)) > desiredTotal {
		return allCommits[:desiredTotal]
	}
	return allCommits
}

// runCommitLog sends the git command to gitserver. It interprets missing
// revision responses and converts them into RevisionNotFoundError.
// It is declared as a variable so that we can swap it out in tests
var runCommitLog = func(ctx context.Context, cmd GitCommand, opt CommitsOptions) ([]*wrappedCommit, error) {
	data, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		data = bytes.TrimSpace(data)
		if isBadObjectErr(string(stderr), opt.Range) {
			return nil, &gitdomain.RevisionNotFoundError{Repo: cmd.Repo(), Spec: opt.Range}
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), data))
	}

	return parseCommitLogOutput(bytes.NewReader(data))
}

func parseCommitLogOutput(r io.Reader) ([]*wrappedCommit, error) {
	commitScanner := bufio.NewScanner(r)
	commitScanner.Split(commitSplitFunc)

	var commits []*wrappedCommit
	for commitScanner.Scan() {
		rawCommit := commitScanner.Bytes()
		parts := bytes.Split(rawCommit, []byte{'\x00'})
		if len(parts) != partsPerCommit {
			return nil, errors.Newf("internal error: expected %d parts, got %d", partsPerCommit, len(parts))
		}

		commit, err := parseCommitFromLog(parts)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func commitSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		// Request more data
		return 0, nil, nil
	}

	// Safety check: ensure we are always starting with a record separator
	if data[0] != '\x1e' {
		return 0, nil, errors.New("internal error: data should always start with an ASCII record separator")
	}

	loc := bytes.IndexByte(data[1:], '\x1e')
	if loc < 0 {
		// We can't find the start of the next record
		if atEOF {
			// If we're at the end of the stream, just return the rest as the last record
			return len(data), data[1:], bufio.ErrFinalToken
		} else {
			// If we're not at the end of the stream, request more data
			return 0, nil, nil
		}
	}
	nextStart := loc + 1 // correct for searching at an offset

	return nextStart, data[1:nextStart], nil
}

type wrappedCommit struct {
	*gitdomain.Commit
	files []string
}

func commitLogArgs(initialArgs []string, opt CommitsOptions) (args []string, err error) {
	if err := checkSpecArgSafety(opt.Range); err != nil {
		return nil, err
	}

	args = initialArgs
	if opt.N != 0 {
		args = append(args, "-n", strconv.FormatUint(uint64(opt.N), 10))
	}
	if opt.Skip != 0 {
		args = append(args, "--skip="+strconv.FormatUint(uint64(opt.Skip), 10))
	}

	if opt.Author != "" {
		args = append(args, "--fixed-strings", "--author="+opt.Author)
	}

	if opt.After != "" {
		args = append(args, "--after="+opt.After)
	}
	if opt.Before != "" {
		args = append(args, "--before="+opt.Before)
	}
	if opt.Reverse {
		args = append(args, "--reverse")
	}
	if opt.DateOrder {
		args = append(args, "--date-order")
	}

	if opt.MessageQuery != "" {
		args = append(args, "--fixed-strings", "--regexp-ignore-case", "--grep="+opt.MessageQuery)
	}

	if opt.Range != "" {
		args = append(args, opt.Range)
	}
	if opt.NameOnly {
		args = append(args, "--name-only")
	}
	if opt.Follow {
		args = append(args, "--follow")
	}
	if opt.Path != "" {
		args = append(args, "--", opt.Path)
	}
	return args, nil
}

// FirstEverCommit returns the first commit ever made to the repository.
func (c *clientImplementor) FirstEverCommit(ctx context.Context, repo api.RepoName) (_ *gitdomain.Commit, err error) {
	ctx, _, endObservation := c.operations.firstEverCommit.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	args := []string{"rev-list", "--reverse", "--date-order", "--max-parents=0", "HEAD"}
	cmd := c.gitCommand(repo, args...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", args, out))
	}
	lines := bytes.TrimSpace(out)
	tokens := bytes.SplitN(lines, []byte("\n"), 2)
	if len(tokens) == 0 {
		return nil, errors.New("FirstEverCommit returned no revisions")
	}
	first := tokens[0]
	id := api.CommitID(bytes.TrimSpace(first))
	return c.GetCommit(ctx, repo, id, ResolveRevisionOptions{NoEnsureRevision: true})
}

// CommitExists determines if the given commit exists in the given repository.
func (c *clientImplementor) CommitExists(ctx context.Context, repo api.RepoName, id api.CommitID) (_ bool, err error) {
	ctx, _, endObservation := c.operations.commitExists.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("commit", string(id)),
		},
	})
	defer endObservation(1, observation.Args{})

	commit, err := c.getCommit(ctx, repo, id, ResolveRevisionOptions{NoEnsureRevision: true})
	if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return commit != nil, nil
}

// CommitsExist determines if the given commits exists in the given repositories. This function returns
// a slice of the same size as the input slice, true indicating that the commit at the symmetric index
// exists.
func (c *clientImplementor) CommitsExist(ctx context.Context, repoCommits []api.RepoCommit) ([]bool, error) {
	commits, err := c.GetCommits(ctx, repoCommits, true)
	if err != nil {
		return nil, err
	}

	exists := make([]bool, len(commits))
	for i, commit := range commits {
		exists[i] = commit != nil
	}

	return exists, nil
}

// GetCommits returns a git commit object describing each of the given repository and commit pairs. This
// function returns a slice of the same size as the input slice. Values in the output slice may be nil if
// their associated repository or commit are unresolvable.
//
// If ignoreErrors is true, then errors arising from any single failed git log operation will cause the
// resulting commit to be nil, but not fail the entire operation.
func (c *clientImplementor) GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool) (_ []*gitdomain.Commit, err error) {
	ctx, _, endObservation := c.operations.getCommits.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			attribute.Int("numRepoCommits", len(repoCommits)),
			attribute.Bool("ignoreErrors", ignoreErrors),
		},
	})
	defer endObservation(1, observation.Args{})

	indexesByRepoCommit := make(map[api.RepoCommit]int, len(repoCommits))
	for i, repoCommit := range repoCommits {
		if err := checkSpecArgSafety(string(repoCommit.CommitID)); err != nil {
			return nil, err
		}

		// Ensure repository names are normalized. If do this in a lower layer, then we may
		// not be able to compare the RepoCommit parameter in the callback below with the
		// input values.
		repoCommits[i].Repo = protocol.NormalizeRepo(repoCommit.Repo)

		// Make it easy to look up the index to populate for a particular RepoCommit value.
		// Note that we use the slice-indexed version as the key, not the local variable, which
		// was not updated in the normalization phase above
		indexesByRepoCommit[repoCommits[i]] = i
	}

	// Create a slice with values populated in the callback defined below. Since the callback
	// may be invoked concurrently inside BatchLog, we need to synchronize writes to this slice
	// with this local mutex.
	commits := make([]*gitdomain.Commit, len(repoCommits))
	var mu sync.Mutex

	callback := func(repoCommit api.RepoCommit, rawResult RawBatchLogResult) error {
		if err := rawResult.Error; err != nil {
			if ignoreErrors {
				// Treat as not-found
				return nil
			}

			return errors.Wrap(err, "failed to perform git log")
		}

		wrappedCommits, err := parseCommitLogOutput(strings.NewReader(rawResult.Stdout))
		if err != nil {
			if ignoreErrors {
				// Treat as not-found
				return nil
			}
			return errors.Wrap(err, "parseCommitLogOutput")
		}
		if len(wrappedCommits) > 1 {
			// Check this prior to filtering commits so that we still log an issue
			// if the user happens to have access one but not the other; a rev being
			// ambiguous here should be a visible issue regardless of permissions.
			return errors.Errorf("git log: expected 1 commit, got %d", len(commits))
		}

		// Enforce sub-repository permissions
		filteredCommits, err := filterCommits(ctx, c.subRepoPermsChecker, wrappedCommits, repoCommit.Repo)
		if err != nil {
			// Note that we don't check ignoreErrors on this condition. When we
			// ignore errors it's to hide an issue with a single git log request on a
			// single shard, which could return an error if that repo is missing, the
			// supplied commit does not exist in the clone, or if the repo is malformed.
			//
			// We don't want to hide unrelated infrastructure errors caused by this
			// method call.
			return errors.Wrap(err, "filterCommits")
		}
		if len(filteredCommits) == 0 {
			// Not found
			return nil
		}

		mu.Lock()
		defer mu.Unlock()
		index := indexesByRepoCommit[repoCommit]
		commits[index] = filteredCommits[0]
		return nil
	}

	opts := BatchLogOptions{
		RepoCommits: repoCommits,
		Format:      logFormatWithoutRefs,
	}
	if err := c.BatchLog(ctx, opts, callback); err != nil {
		return nil, errors.Wrap(err, "gitserver.BatchLog")
	}

	return commits, nil
}

// Head determines the tip commit of the default branch for the given repository.
// If no HEAD revision exists for the given repository (which occurs with empty
// repositories), a false-valued flag is returned along with a nil error and
// empty revision.
func (c *clientImplementor) Head(ctx context.Context, repo api.RepoName) (_ string, revisionExists bool, err error) {
	ctx, _, endObservation := c.operations.head.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	cmd := c.gitCommand(repo, "rev-parse", "HEAD")

	out, err := cmd.Output(ctx)
	if err != nil {
		return checkError(err)
	}
	commitID := string(out)
	if authz.SubRepoEnabled(c.subRepoPermsChecker) {
		if _, err := c.GetCommit(ctx, repo, api.CommitID(commitID), ResolveRevisionOptions{
			// We expect the HEAD reference to not be broken, let's not ensure the
			// revision here.
			NoEnsureRevision: true,
		}); err != nil {
			return checkError(err)
		}
	}

	return commitID, true, nil
}

func checkError(err error) (string, bool, error) {
	if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
		err = nil
	}
	return "", false, err
}

const (
	partsPerCommit = 10 // number of \x00-separated fields per commit

	// This format string has 10 parts:
	//  1) oid
	//  2) author name
	//  3) author email
	//  4) author time
	//  5) committer name
	//  6) committer email
	//  7) committer time
	//  8) message body
	//  9) parent hashes
	// 10) modified files (optional)
	//
	// Each commit starts with an ASCII record separator byte (0x1E), and
	// each field of the commit is separated by a null byte (0x00).
	//
	// Refs are slow, and are intentionally not included because they are usually not needed.
	logFormatWithoutRefs = "--format=format:%x1e%H%x00%aN%x00%aE%x00%at%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00"
)

// parseCommitFromLog parses the next commit from data and returns the commit and the remaining
// data. The data arg is a byte array that contains NUL-separated log fields as formatted by
// logFormatFlag.
func parseCommitFromLog(parts [][]byte) (*wrappedCommit, error) {
	// log outputs are newline separated, so all but the 1st commit ID part
	// has an erroneous leading newline.
	parts[0] = bytes.TrimPrefix(parts[0], []byte{'\n'})
	commitID := api.CommitID(parts[0])

	authorTime, err := strconv.ParseInt(string(parts[3]), 10, 64)
	if err != nil {
		return nil, errors.Errorf("parsing git commit author time: %s", err)
	}
	committerTime, err := strconv.ParseInt(string(parts[6]), 10, 64)
	if err != nil {
		return nil, errors.Errorf("parsing git commit committer time: %s", err)
	}

	var parents []api.CommitID
	if parentPart := parts[8]; len(parentPart) > 0 {
		parentIDs := bytes.Split(parentPart, []byte{' '})
		parents = make([]api.CommitID, len(parentIDs))
		for i, id := range parentIDs {
			parents[i] = api.CommitID(id)
		}
	}

	fileNames := strings.Split(string(bytes.TrimSpace(parts[9])), "\n")

	return &wrappedCommit{
		Commit: &gitdomain.Commit{
			ID:        commitID,
			Author:    gitdomain.Signature{Name: string(parts[1]), Email: string(parts[2]), Date: time.Unix(authorTime, 0).UTC()},
			Committer: &gitdomain.Signature{Name: string(parts[4]), Email: string(parts[5]), Date: time.Unix(committerTime, 0).UTC()},
			Message:   gitdomain.Message(strings.TrimSuffix(string(parts[7]), "\n")),
			Parents:   parents,
		}, files: fileNames,
	}, nil
}

// BranchesContaining returns a map from branch names to branch tip hashes for
// each branch containing the given commit.
func (c *clientImplementor) BranchesContaining(ctx context.Context, repo api.RepoName, commit api.CommitID) (_ []string, err error) {
	ctx, _, endObservation := c.operations.branchesContaining.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("commit", string(commit)),
		},
	})
	defer endObservation(1, observation.Args{})

	if authz.SubRepoEnabled(c.subRepoPermsChecker) {
		// GetCommit to validate that the user has permissions to access it.
		if _, err := c.GetCommit(ctx, repo, commit, ResolveRevisionOptions{
			// Don't ensure the revision here, the branch command below will fail
			// anyways.
			NoEnsureRevision: true,
		}); err != nil {
			return nil, err
		}
	}
	cmd := c.gitCommand(repo, "branch", "--contains", string(commit), "--format", "%(refname)")

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return parseBranchesContaining(strings.Split(string(out), "\n")), nil
}

var refReplacer = strings.NewReplacer("refs/heads/", "", "refs/tags/", "")

func parseBranchesContaining(lines []string) []string {
	names := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = refReplacer.Replace(line)
		names = append(names, line)
	}
	sort.Strings(names)

	return names
}

// RefDescriptions returns a map from commits to descriptions of the tip of each
// branch and tag of the given repository.
func (c *clientImplementor) RefDescriptions(ctx context.Context, repo api.RepoName, gitObjs ...string) (_ map[string][]gitdomain.RefDescription, err error) {
	ctx, _, endObservation := c.operations.refDescriptions.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.Int("objects", len(gitObjs)),
		},
	})
	defer endObservation(1, observation.Args{})

	f := func(refPrefix string) (map[string][]gitdomain.RefDescription, error) {
		format := strings.Join([]string{
			derefField("objectname"),
			"%(refname)",
			"%(HEAD)",
			derefField("creatordate:iso8601-strict"),
		}, "%00")

		args := make([]string, 0, len(gitObjs)+3)
		args = append(args, "for-each-ref", "--format="+format, refPrefix)

		for _, obj := range gitObjs {
			args = append(args, "--points-at="+obj)
		}

		cmd := c.gitCommand(repo, args...)

		out, err := cmd.CombinedOutput(ctx)
		if err != nil {
			return nil, err
		}

		return parseRefDescriptions(out)
	}

	aggregate := make(map[string][]gitdomain.RefDescription)
	for prefix := range refPrefixes {
		descriptions, err := f(prefix)
		if err != nil {
			return nil, err
		}
		for commit, descs := range descriptions {
			aggregate[commit] = append(aggregate[commit], descs...)
		}
	}

	if authz.SubRepoEnabled(c.subRepoPermsChecker) {
		return c.filterRefDescriptions(ctx, repo, aggregate), nil
	}
	return aggregate, nil
}

func derefField(field string) string {
	return "%(if)%(*" + field + ")%(then)%(*" + field + ")%(else)%(" + field + ")%(end)"
}

func (c *clientImplementor) filterRefDescriptions(ctx context.Context,
	repo api.RepoName,
	refDescriptions map[string][]gitdomain.RefDescription,
) map[string][]gitdomain.RefDescription {
	filtered := make(map[string][]gitdomain.RefDescription, len(refDescriptions))
	for commitID, descriptions := range refDescriptions {
		_, err := c.GetCommit(ctx, repo, api.CommitID(commitID), ResolveRevisionOptions{
			// The commits are returned from git already, so they must exist.
			NoEnsureRevision: true,
		})
		if !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			filtered[commitID] = descriptions
		}
	}
	return filtered
}

var refPrefixes = map[string]gitdomain.RefType{
	"refs/heads/": gitdomain.RefTypeBranch,
	"refs/tags/":  gitdomain.RefTypeTag,
}

// parseRefDescriptions converts the output of the for-each-ref command in the RefDescriptions
// method to a map from commits to RefDescription objects. The output is expected to be a series
// of lines each conforming to  `%(objectname)%00%(refname)%00%(HEAD)%00%(creatordate)`, where
//
// - %(objectname) is the 40-character revhash
// - %(refname) is the name of the tag or branch (prefixed with refs/heads/ or ref/tags/)
// - %(HEAD) is `*` if the branch is the default branch (and whitesace otherwise)
// - %(creatordate) is the ISO-formatted date the object was created
func parseRefDescriptions(out []byte) (map[string][]gitdomain.RefDescription, error) {
	refDescriptions := make(map[string][]gitdomain.RefDescription, bytes.Count(out, []byte("\n")))

	lr := byteutils.NewLineReader(out)

lineLoop:
	for lr.Scan() {
		line := bytes.TrimSpace(lr.Line())
		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte("\x00"), 4)
		if len(parts) != 4 {
			return nil, errors.Errorf(`unexpected output from git for-each-ref %q`, string(line))
		}

		commit := string(parts[0])
		isDefaultBranch := string(parts[2]) == "*"

		var name string
		var refType gitdomain.RefType
		for prefix, typ := range refPrefixes {
			if strings.HasPrefix(string(parts[1]), prefix) {
				name = string(parts[1])[len(prefix):]
				refType = typ
				break
			}
		}
		if refType == gitdomain.RefTypeUnknown {
			return nil, errors.Errorf(`unexpected output from git for-each-ref "%s"`, line)
		}

		var (
			createdDatePart = string(parts[3])
			createdDatePtr  *time.Time
		)
		// Some repositories attach tags to non-commit objects, such as trees. In such a situation, one
		// cannot deference the tag to obtain the commit it points to, and there is no associated creatordate.
		if createdDatePart != "" {
			createdDate, err := time.Parse(time.RFC3339, createdDatePart)
			if err != nil {
				return nil, errors.Errorf(`unexpected output from git for-each-ref (bad date format) "%s"`, line)
			}
			createdDatePtr = &createdDate
		}

		// Check for duplicates before adding it to the slice
		for _, candidate := range refDescriptions[commit] {
			if candidate.Name == name && candidate.Type == refType && candidate.IsDefaultBranch == isDefaultBranch {
				continue lineLoop
			}
		}

		refDescriptions[commit] = append(refDescriptions[commit], gitdomain.RefDescription{
			Name:            name,
			Type:            refType,
			IsDefaultBranch: isDefaultBranch,
			CreatedDate:     createdDatePtr,
		})
	}

	return refDescriptions, nil
}

// CommitDate returns the time that the given commit was committed. If the given
// revision does not exist, a false-valued flag is returned along with a nil
// error and zero-valued time.
func (c *clientImplementor) CommitDate(ctx context.Context, repo api.RepoName, commit api.CommitID) (_ string, _ time.Time, revisionExists bool, err error) {
	ctx, _, endObservation := c.operations.commitDate.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
			attribute.String("commit", string(commit)),
		},
	})
	defer endObservation(1, observation.Args{})

	if authz.SubRepoEnabled(c.subRepoPermsChecker) {
		// GetCommit to validate that the user has permissions to access it.
		if _, err := c.GetCommit(ctx, repo, commit, ResolveRevisionOptions{
			// The show command below will fail anyways, so no need to ensure
			// that the revision exists here.
			NoEnsureRevision: true,
		}); err != nil {
			return "", time.Time{}, false, nil
		}
	}

	cmd := c.gitCommand(repo, "show", "-s", "--format=%H:%cI", string(commit))

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			err = nil
		}
		return "", time.Time{}, false, err
	}
	outs := string(out)

	line := strings.TrimSpace(outs)
	if line == "" {
		return "", time.Time{}, false, nil
	}

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", time.Time{}, false, errors.Errorf(`unexpected output from git show "%s"`, line)
	}

	duration, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return "", time.Time{}, false, errors.Errorf(`unexpected output from git show (bad date format) "%s"`, line)
	}

	return parts[0], duration, true, nil
}

type ArchiveFormat string

const (
	// ArchiveFormatZip indicates a zip archive is desired.
	ArchiveFormatZip ArchiveFormat = "zip"

	// ArchiveFormatTar indicates a tar archive is desired.
	ArchiveFormatTar ArchiveFormat = "tar"
)

// ArchiveReader streams back the file contents of an archived git repo.
func (c *clientImplementor) ArchiveReader(
	ctx context.Context,
	repo api.RepoName,
	options ArchiveOptions,
) (_ io.ReadCloser, err error) {
	// TODO: this does not capture the lifetime of the request because we return a reader
	ctx, _, endObservation := c.operations.archiveReader.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: append(
			[]attribute.KeyValue{repo.Attr()},
			options.Attrs()...,
		),
	})
	defer endObservation(1, observation.Args{})

	if authz.SubRepoEnabled(c.subRepoPermsChecker) {
		if enabled, err := authz.SubRepoEnabledForRepo(ctx, c.subRepoPermsChecker, repo); err != nil {
			return nil, errors.Wrap(err, "sub-repo permissions check:")
		} else if enabled {
			return nil, errors.New("archiveReader invoked for a repo with sub-repo permissions")
		}
	}

	if ClientMocks.Archive != nil {
		return ClientMocks.Archive(ctx, repo, options)
	}

	// Check that ctx is not expired.
	if err := ctx.Err(); err != nil {
		deadlineExceededCounter.Inc()
		return nil, err
	}

	if conf.IsGRPCEnabled(ctx) {
		client, err := c.clientSource.ClientForRepo(ctx, c.userAgent, repo)
		if err != nil {
			return nil, err
		}

		req := options.ToProto(string(repo)) // HACK: ArchiveOptions doesn't have a repository here, so we have to add it ourselves.

		ctx, cancel := context.WithCancel(ctx)

		stream, err := client.Archive(ctx, req)
		if err != nil {
			cancel()
			return nil, err
		}

		// first message from the gRPC stream needs to be read to check for errors before continuing
		// to read the rest of the stream. If the first message is an error, we cancel the stream
		// and return the error.
		//
		// This is necessary to provide parity between the REST and gRPC implementations of
		// ArchiveReader. Users of cli.ArchiveReader may assume error handling occurs immediately,
		// as is the case with the HTTP implementation where errors are returned as soon as the
		// function returns. gRPC is asynchronous, so we have to start consuming messages from
		// the stream to see any errors from the server. Reading the first message ensures we
		// handle any errors synchronously, similar to the HTTP implementation.

		firstMessage, firstError := stream.Recv()
		if firstError != nil {
			// Hack: The ArchiveReader.Read() implementation handles surfacing the
			// any "revision not found" errors returned from the invoked git binary.
			//
			// In order to maintainparity with the HTTP API, we return this error in the ArchiveReader.Read() method
			// instead of returning it immediately.

			// We return early only if this isn't a revision not found error.

			err := convertGRPCErrorToGitDomainError(firstError)

			var cse *CommandStatusError
			if !errors.As(err, &cse) || !isRevisionNotFound(cse.Stderr) {
				cancel()
				return nil, convertGRPCErrorToGitDomainError(err)
			}
		}

		firstMessageRead := false

		// Create a reader to read from the gRPC stream.
		r := streamio.NewReader(func() ([]byte, error) {
			// Check if we've read the first message yet. If not, read it and return.
			if !firstMessageRead {
				firstMessageRead = true

				if firstError != nil {
					return nil, firstError
				}

				return firstMessage.GetData(), nil
			}

			// Receive the next message from the stream.
			msg, err := stream.Recv()
			if err != nil {
				return nil, convertGRPCErrorToGitDomainError(err)
			}

			// Return the data from the received message.
			return msg.GetData(), nil
		})

		return &archiveReader{
			base: &readCloseWrapper{r: r, closeFn: cancel},
			repo: repo,
			spec: options.Treeish,
		}, nil

	} else {
		// Fall back to http request
		u := c.archiveURL(ctx, repo, options)
		resp, err := c.do(ctx, repo, u.String(), nil)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return &archiveReader{
				base: &cmdReader{
					rc:      resp.Body,
					trailer: resp.Trailer,
				},
				repo: repo,
				spec: options.Treeish,
			}, nil
		case http.StatusNotFound:
			var payload protocol.NotFoundPayload
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				resp.Body.Close()
				return nil, err
			}
			resp.Body.Close()
			return nil, &badRequestError{
				error: &gitdomain.RepoNotExistError{
					Repo:            repo,
					CloneInProgress: payload.CloneInProgress,
					CloneProgress:   payload.CloneProgress,
				},
			}
		default:
			resp.Body.Close()
			return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	}
}

func addNameOnly(opt CommitsOptions, checker authz.SubRepoPermissionChecker) CommitsOptions {
	if authz.SubRepoEnabled(checker) {
		// If sub-repo permissions enabled, must fetch files modified w/ commits to determine if user has access to view this commit
		opt.NameOnly = true
	}
	return opt
}

// BranchesOptions specifies options for the list of branches returned by
// (Repository).Branches.
type BranchesOptions struct {
	// MergedInto will cause the returned list to be restricted to only
	// branches that were merged into this branch name.
	MergedInto string `json:"MergedInto,omitempty" url:",omitempty"`
	// IncludeCommit controls whether complete commit information is included.
	IncludeCommit bool `json:"IncludeCommit,omitempty" url:",omitempty"`
	// BehindAheadBranch specifies a branch name. If set to something other than blank
	// string, then each returned branch will include a behind/ahead commit counts
	// information against the specified base branch. If left blank, then branches will
	// not include that information and their Counts will be nil.
	BehindAheadBranch string `json:"BehindAheadBranch,omitempty" url:",omitempty"`
	// ContainsCommit filters the list of branches to only those that
	// contain a specific commit ID (if set).
	ContainsCommit string `json:"ContainsCommit,omitempty" url:",omitempty"`
}

func (bo *BranchesOptions) Attrs() (res []attribute.KeyValue) {
	if bo.MergedInto != "" {
		res = append(res, attribute.String("mergedInto", bo.MergedInto))
	}
	res = append(res, attribute.Bool("includeCommit", bo.IncludeCommit))

	if bo.BehindAheadBranch != "" {
		res = append(res, attribute.String("behindAheadBranch", bo.BehindAheadBranch))
	}

	if bo.ContainsCommit != "" {
		res = append(res, attribute.String("containsCommit", bo.ContainsCommit))
	}

	return res
}

// branchFilter is a filter for branch names.
// If not empty, only contained branch names are allowed. If empty, all names are allowed.
// The map should be made so it's not nil.
type branchFilter map[string]struct{}

// allows will return true if the current filter set-up validates against
// the passed string. If there are no filters, all strings pass.
func (f branchFilter) allows(name string) bool {
	if len(f) == 0 {
		return true
	}
	_, ok := f[name]
	return ok
}

// add adds a slice of strings to the filter.
func (f branchFilter) add(list []string) {
	for _, l := range list {
		f[l] = struct{}{}
	}
}

// ListBranches returns a list of all branches in the repository.
func (c *clientImplementor) ListBranches(ctx context.Context, repo api.RepoName, opt BranchesOptions) (_ []*gitdomain.Branch, err error) {
	ctx, _, endObservation := c.operations.listBranches.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: append(
			[]attribute.KeyValue{repo.Attr()},
			opt.Attrs()...,
		),
	})
	defer endObservation(1, observation.Args{})

	f := make(branchFilter)
	if opt.MergedInto != "" {
		b, err := c.branches(ctx, repo, "--merged", opt.MergedInto)
		if err != nil {
			return nil, err
		}
		f.add(b)
	}
	if opt.ContainsCommit != "" {
		b, err := c.branches(ctx, repo, "--contains="+opt.ContainsCommit)
		if err != nil {
			return nil, err
		}
		f.add(b)
	}

	refs, err := c.showRef(ctx, repo, "--heads")
	if err != nil {
		return nil, err
	}

	var branches []*gitdomain.Branch
	for _, ref := range refs {
		name := strings.TrimPrefix(ref.Name, "refs/heads/")
		if !f.allows(name) {
			continue
		}

		branch := &gitdomain.Branch{Name: name, Head: ref.CommitID}
		if opt.IncludeCommit {
			branch.Commit, err = c.GetCommit(ctx, repo, ref.CommitID, ResolveRevisionOptions{
				// The commitID comes from git, so surely it knows about it and
				// there's no need to ensure the revision exists.
				NoEnsureRevision: true,
			})
			if err != nil {
				return nil, err
			}
		}
		if opt.BehindAheadBranch != "" {
			branch.Counts, err = c.GetBehindAhead(ctx, repo, "refs/heads/"+opt.BehindAheadBranch, "refs/heads/"+name)
			if err != nil {
				return nil, err
			}
		}
		branches = append(branches, branch)
	}
	return branches, nil
}

// branches runs the `git branch` command followed by the given arguments and
// returns the list of branches if successful.
func (c *clientImplementor) branches(ctx context.Context, repo api.RepoName, args ...string) ([]string, error) {
	cmd := c.gitCommand(repo, append([]string{"branch"}, args...)...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.Errorf("exec %v in %s failed: %v (output follows)\n\n%s", cmd.Args(), cmd.Repo(), err, out)
	}
	lines := strings.Split(string(out), "\n")
	lines = lines[:len(lines)-1]
	branches := make([]string, len(lines))
	for i, line := range lines {
		branches[i] = line[2:]
	}
	return branches, nil
}

type byteSlices [][]byte

func (p byteSlices) Len() int           { return len(p) }
func (p byteSlices) Less(i, j int) bool { return bytes.Compare(p[i], p[j]) < 0 }
func (p byteSlices) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// ListRefs returns a list of all refs in the repository.
func (c *clientImplementor) ListRefs(ctx context.Context, repo api.RepoName) (_ []gitdomain.Ref, err error) {
	ctx, _, endObservation := c.operations.listRefs.With(ctx, &err, observation.Args{
		MetricLabelValues: []string{c.scope},
		Attrs: []attribute.KeyValue{
			repo.Attr(),
		},
	})
	defer endObservation(1, observation.Args{})

	return c.showRef(ctx, repo)
}

func (c *clientImplementor) showRef(ctx context.Context, repo api.RepoName, args ...string) ([]gitdomain.Ref, error) {
	cmdArgs := append([]string{"show-ref"}, args...)
	cmd := c.gitCommand(repo, cmdArgs...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if gitdomain.IsRepoNotExist(err) {
			return nil, err
		}
		// Exit status of 1 and no output means there were no
		// results. This is not a fatal error.
		if cmd.ExitStatus() == 1 && len(out) == 0 {
			return nil, nil
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}

	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
	lines := bytes.Split(out, []byte("\n"))
	sort.Sort(byteSlices(lines)) // sort for consistency
	refs := make([]gitdomain.Ref, len(lines))
	for i, line := range lines {
		if len(line) <= 41 {
			return nil, errors.New("unexpectedly short (<=41 bytes) line in `git show-ref ...` output")
		}
		id := line[:40]
		name := line[41:]
		refs[i] = gitdomain.Ref{Name: string(name), CommitID: api.CommitID(id)}
	}
	return refs, nil
}

// rel strips the leading "/" prefix from the path string, effectively turning
// an absolute path into one relative to the root directory. A path that is just
// "/" is treated specially, returning just ".".
//
// The elements in a file path are separated by slash ('/', U+002F) characters,
// regardless of host operating system convention.
func rel(path string) string {
	if path == "/" {
		return "."
	}
	return strings.TrimPrefix(path, "/")
}
