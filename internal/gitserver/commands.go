package gitserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
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
	"github.com/grafana/regexp"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/vcs/util"
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
}

// Diff returns an iterator that can be used to access the diff between two
// commits on a per-file basis. The iterator must be closed with Close when no
// longer required.
func (c *ClientImplementor) Diff(ctx context.Context, opts DiffOptions) (*DiffFileIterator, error) {
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

	rdr, err := c.execReader(ctx, opts.Repo, []string{
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
	})
	if err != nil {
		return nil, errors.Wrap(err, "executing git diff")
	}

	return &DiffFileIterator{
		rdr:  rdr,
		mfdr: diff.NewMultiFileDiffReader(rdr),
	}, nil
}

type DiffFileIterator struct {
	rdr  io.ReadCloser
	mfdr *diff.MultiFileDiffReader
}

func (i *DiffFileIterator) Close() error {
	return i.rdr.Close()
}

// Next returns the next file diff. If no more diffs are available, the diff
// will be nil and the error will be io.EOF.
func (i *DiffFileIterator) Next() (*diff.FileDiff, error) {
	return i.mfdr.ReadFile()
}

// ShortLogOptions contains options for (Repository).ShortLog.
type ShortLogOptions struct {
	Range string // the range for which stats will be fetched
	After string // the date after which to collect commits
	Path  string // compute stats for commits that touch this path
}

func (c *ClientImplementor) ShortLog(ctx context.Context, repo api.RepoName, opt ShortLogOptions) ([]*gitdomain.PersonCount, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: ShortLog")
	span.SetTag("Opt", opt)
	defer span.Finish()

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
	cmd := c.GitCommand(repo, args...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.Errorf("exec `git shortlog -s -n -e` failed: %v", err)
	}
	return parseShortLog(out)
}

// execReader executes an arbitrary `git` command (`git [args...]`) and returns a
// reader connected to its stdout.
//
// execReader should NOT be exported. We want to limit direct git calls to this
// package.
func (c *ClientImplementor) execReader(ctx context.Context, repo api.RepoName, args []string) (io.ReadCloser, error) {
	if Mocks.ExecReader != nil {
		return Mocks.ExecReader(args)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: ExecReader")
	span.SetTag("args", args)
	defer span.Finish()

	if !gitdomain.IsAllowedGitCmd(c.logger, args) {
		return nil, errors.Errorf("command failed: %v is not a allowed git command", args)
	}
	cmd := c.GitCommand(repo, args...)
	return cmd.StdoutReader(ctx)
}

// logEntryPattern is the regexp pattern that matches entries in the output of the `git shortlog
// -sne` command.
var logEntryPattern = lazyregexp.New(`^\s*([0-9]+)\s+(.*)$`)

func parseShortLog(out []byte) ([]*gitdomain.PersonCount, error) {
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, nil
	}
	lines := bytes.Split(out, []byte{'\n'})
	results := make([]*gitdomain.PersonCount, len(lines))
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
		results[i] = &gitdomain.PersonCount{
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
// 	foo@gmail.com <foo@gmail.com>
//
// As a valid name, whereas mail.ParseAddress would return an error:
//
// 	mail: expected single address, got "<foo@gmail.com>"
//
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

func stableTimeRepr(t time.Time) string {
	s, _ := t.MarshalText()
	return string(s)
}

func (opts *CommitGraphOptions) LogFields() []log.Field {
	var since string
	if opts.Since != nil {
		since = stableTimeRepr(*opts.Since)
	} else {
		since = stableTimeRepr(time.Unix(0, 0))
	}

	return []log.Field{
		log.String("commit", opts.Commit),
		log.Int("limit", opts.Limit),
		log.Bool("allrefs", opts.AllRefs),
		log.String("since", since),
	}
}

// CommitGraph returns the commit graph for the given repository as a mapping
// from a commit to its parents. If a commit is supplied, the returned graph will
// be rooted at the given commit. If a non-zero limit is supplied, at most that
// many commits will be returned.
func (c *ClientImplementor) CommitGraph(ctx context.Context, repo api.RepoName, opts CommitGraphOptions) (_ *gitdomain.CommitGraph, err error) {
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

	cmd := c.GitCommand(repo, args...)

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return gitdomain.ParseCommitGraph(strings.Split(string(out), "\n")), nil
}

// DevNullSHA 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t
// tree /dev/null`, which is used as the base when computing the `git diff` of
// the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func (c *ClientImplementor) DiffPath(ctx context.Context, repo api.RepoName, sourceCommit, targetCommit, path string, checker authz.SubRepoPermissionChecker) ([]*diff.Hunk, error) {
	a := actor.FromContext(ctx)
	if hasAccess, err := authz.FilterActorPath(ctx, checker, a, repo, path); err != nil {
		return nil, err
	} else if !hasAccess {
		return nil, os.ErrNotExist
	}
	reader, err := c.execReader(ctx, repo, []string{"diff", sourceCommit, targetCommit, "--", path})
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
func (c *ClientImplementor) DiffSymbols(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) ([]byte, error) {
	command := c.GitCommand(repo, "diff", "-z", "--name-status", "--no-renames", string(commitA), string(commitB))
	return command.Output(ctx)
}

// ReadDir reads the contents of the named directory at commit.
func (c *ClientImplementor) ReadDir(
	ctx context.Context,
	db database.DB,
	checker authz.SubRepoPermissionChecker,
	repo api.RepoName,
	commit api.CommitID,
	path string,
	recurse bool,
) ([]fs.FileInfo, error) {
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
	files, err := c.lsTree(ctx, repo, commit, path, recurse)

	if err != nil || !authz.SubRepoEnabled(checker) {
		return files, err
	}

	a := actor.FromContext(ctx)
	filtered, filteringErr := authz.FilterActorFileInfos(ctx, checker, a, repo, files)
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
func (c *ClientImplementor) lsTree(
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

// LStat returns a FileInfo describing the named file at commit. If the file is a symbolic link, the
// returned FileInfo describes the symbolic link.  lStat makes no attempt to follow the link.
// TODO(sashaostrikov): make private when git.Stat is moved here as well
func (c *ClientImplementor) LStat(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
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
		obj, err := c.GetObject(ctx, repo, string(commit)+"^{tree}")
		if err != nil {
			return nil, err
		}
		return &util.FileInfo{Mode_: os.ModeDir, Sys_: objectInfo(obj.ID)}, nil
	}

	fis, err := c.lsTree(ctx, repo, commit, path, false)
	if err != nil {
		return nil, err
	}
	if len(fis) == 0 {
		return nil, &os.PathError{Op: "ls-tree", Path: path, Err: os.ErrNotExist}
	}

	if !authz.SubRepoEnabled(checker) {
		return fis[0], nil
	}
	// Applying sub-repo permissions
	a := actor.FromContext(ctx)
	include, filteringErr := authz.FilterActorFileInfo(ctx, checker, a, repo, fis[0])
	if include && filteringErr == nil {
		return fis[0], nil
	} else {
		if filteringErr != nil {
			err = errors.Wrap(err, "filtering paths")
		} else {
			err = &os.PathError{Op: "ls-tree", Path: path, Err: os.ErrNotExist}
		}
		return nil, err
	}
}

func (c *ClientImplementor) lsTreeUncached(ctx context.Context, repo api.RepoName, commit api.CommitID, path string, recurse bool) ([]fs.FileInfo, error) {
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
	cmd := c.GitCommand(repo, args...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contains(out, []byte("exists on disk, but not in")) {
			return nil, &os.PathError{Op: "ls-tree", Path: filepath.ToSlash(path), Err: os.ErrNotExist}
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
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
			cmd := c.GitCommand(repo, "show", fmt.Sprintf("%s:.gitmodules", commit))
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

func decodeOID(sha string) (gitdomain.OID, error) {
	oidBytes, err := hex.DecodeString(sha)
	if err != nil {
		return gitdomain.OID{}, err
	}
	var oid gitdomain.OID
	copy(oid[:], oidBytes)
	return oid, nil
}

func (c *ClientImplementor) LogReverseEach(repo string, commit string, n int, onLogEntry func(entry gitdomain.LogEntry) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	command := c.GitCommand(api.RepoName(repo), gitdomain.LogReverseArgs(n, commit)...)

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
	NewestCommit api.CommitID `json:",omitempty" url:",omitempty"`
	OldestCommit api.CommitID `json:",omitempty" url:",omitempty"` // or "" for the root commit

	StartLine int `json:",omitempty" url:",omitempty"` // 1-indexed start byte (or 0 for beginning of file)
	EndLine   int `json:",omitempty" url:",omitempty"` // 1-indexed end byte (or 0 for end of file)
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

// BlameFile returns Git blame information about a file.
func (c *ClientImplementor) BlameFile(ctx context.Context, repo api.RepoName, path string, opt *BlameOptions, checker authz.SubRepoPermissionChecker) ([]*Hunk, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: BlameFile")
	span.SetTag("repo", repo)
	span.SetTag("path", path)
	span.SetTag("opt", opt)
	defer span.Finish()
	return blameFileCmd(ctx, c.gitserverGitCommandFunc(repo), path, opt, repo, checker)
}

func blameFileCmd(ctx context.Context, command gitCommandFunc, path string, opt *BlameOptions, repo api.RepoName, checker authz.SubRepoPermissionChecker) ([]*Hunk, error) {
	a := actor.FromContext(ctx)
	if hasAccess, err := authz.FilterActorPath(ctx, checker, a, repo, path); err != nil || !hasAccess {
		return nil, err
	}
	if opt == nil {
		opt = &BlameOptions{}
	}
	if opt.OldestCommit != "" {
		return nil, errors.Errorf("OldestCommit not implemented")
	}
	if err := checkSpecArgSafety(string(opt.NewestCommit)); err != nil {
		return nil, err
	}
	if err := checkSpecArgSafety(string(opt.OldestCommit)); err != nil {
		return nil, err
	}

	args := []string{"blame", "-w", "--porcelain"}
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

	commits := make(map[string]gitdomain.Commit)
	filenames := make(map[string]string)
	hunks := make([]*Hunk, 0)
	remainingLines := strings.Split(string(out[:len(out)-1]), "\n")
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

func (c *ClientImplementor) gitserverGitCommandFunc(repo api.RepoName) gitCommandFunc {
	return func(args []string) GitCommand {
		return c.GitCommand(repo, args...)
	}
}

// gitCommandFunc is a func that creates a new executable Git command.
type gitCommandFunc func(args []string) GitCommand

// IsAbsoluteRevision checks if the revision is a git OID SHA string.
//
// Note: This doesn't mean the SHA exists in a repository, nor does it mean it
// isn't a ref. Git allows 40-char hexadecimal strings to be references.
func IsAbsoluteRevision(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(('0' <= r && r <= '9') ||
			('a' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return false
		}
	}
	return true
}

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
func (c *ClientImplementor) ResolveRevision(ctx context.Context, repo api.RepoName, spec string, opt ResolveRevisionOptions) (api.CommitID, error) {
	if Mocks.ResolveRevision != nil {
		return Mocks.ResolveRevision(spec, opt)
	}

	labelEnsureRevisionValue := "true"
	if opt.NoEnsureRevision {
		labelEnsureRevisionValue = "false"
	}
	resolveRevisionCounter.WithLabelValues(labelEnsureRevisionValue).Inc()

	span, ctx := ot.StartSpanFromContext(ctx, "Git: ResolveRevision")
	span.SetTag("Spec", spec)
	span.SetTag("Opt", fmt.Sprintf("%+v", opt))
	defer span.Finish()

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

	cmd := c.GitCommand(repo, "rev-parse", spec)
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
	if !IsAbsoluteRevision(string(commit)) {
		if commit == "HEAD" {
			// We don't verify the existence of HEAD (see above comments), but
			// if HEAD doesn't point to anything git just returns `HEAD` as the
			// output of rev-parse. An example where this occurs is an empty
			// repository.
			return "", &gitdomain.RevisionNotFoundError{Repo: cmd.Repo(), Spec: spec}
		}
		return "", gitdomain.BadCommitError{Spec: spec, Commit: commit, Repo: cmd.Repo()}
	}
	return commit, nil
}

// LsFiles returns the output of `git ls-files`
func (c *ClientImplementor) LsFiles(ctx context.Context, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, pathspecs ...Pathspec) ([]string, error) {
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

	cmd := c.GitCommand(repo, args...)
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
func (c *ClientImplementor) ListFiles(ctx context.Context, repo api.RepoName, commit api.CommitID, pattern *regexp.Regexp, checker authz.SubRepoPermissionChecker) (_ []string, err error) {
	cmd := c.GitCommand(repo, "ls-tree", "--name-only", "-r", string(commit), "--")

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
func (c *ClientImplementor) ListDirectoryChildren(
	ctx context.Context,
	checker authz.SubRepoPermissionChecker,
	repo api.RepoName,
	commit api.CommitID,
	dirnames []string,
) (map[string][]string, error) {
	args := []string{"ls-tree", "--name-only", string(commit), "--"}
	args = append(args, cleanDirectoriesForLsTree(dirnames)...)
	cmd := c.GitCommand(repo, args...)

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

// ListTags returns a list of all tags in the repository. If commitObjs is non-empty, only all tags pointing at those commits are returned.
func (c *ClientImplementor) ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) ([]*gitdomain.Tag, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: Tags")
	defer span.Finish()

	// Support both lightweight tags and tag objects. For creatordate, use an %(if) to prefer the
	// taggerdate for tag objects, otherwise use the commit's committerdate (instead of just always
	// using committerdate).
	args := []string{"tag", "--list", "--sort", "-creatordate", "--format", "%(if)%(*objectname)%(then)%(*objectname)%(else)%(objectname)%(end)%00%(refname:short)%00%(if)%(creatordate:unix)%(then)%(creatordate:unix)%(else)%(*creatordate:unix)%(end)"}

	for _, commit := range commitObjs {
		args = append(args, "--points-at", commit)
	}

	cmd := c.GitCommand(repo, args...)
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
// currently at from the given repository.
//
// If the repository is empty or currently being cloned, empty values and no
// error are returned.
func (c *ClientImplementor) GetDefaultBranch(ctx context.Context, repo api.RepoName) (refName string, commit api.CommitID, err error) {
	if Mocks.GetDefaultBranch != nil {
		return Mocks.GetDefaultBranch(repo)
	}
	return c.getDefaultBranch(ctx, repo, false)
}

// GetDefaultBranchShort returns the short name of the default branch for the
// given repository and the commit it's currently at. A short name would return
// something like `main` instead of `refs/heads/main`.
//
// If the repository is empty or currently being cloned, empty values and no
// error are returned.
func (c *ClientImplementor) GetDefaultBranchShort(ctx context.Context, repo api.RepoName) (refName string, commit api.CommitID, err error) {
	if Mocks.GetDefaultBranchShort != nil {
		return Mocks.GetDefaultBranchShort(repo)
	}
	return c.getDefaultBranch(ctx, repo, true)
}

// getDefaultBranch returns the name of the default branch and the commit it's
// currently at from the given repository.
//
// If the repository is empty or currently being cloned, empty values and no
// error are returned.
func (c *ClientImplementor) getDefaultBranch(ctx context.Context, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error) {
	args := []string{"symbolic-ref", "HEAD"}
	if short {
		args = append(args, "--short")
	}
	refBytes, _, exitCode, err := c.execSafe(ctx, repo, args)
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

// execSafe executes a Git subcommand iff it is allowed according to a allowlist.
//
// An error is only returned when there is a failure unrelated to the actual
// command being executed. If the executed command exits with a nonzero exit
// code, err == nil. This is similar to how http.Get returns a nil error for HTTP
// non-2xx responses.
//
// execSafe should NOT be exported. We want to limit direct git calls to this
// package.
func (c *ClientImplementor) execSafe(ctx context.Context, repo api.RepoName, params []string) (stdout, stderr []byte, exitCode int, err error) {
	if Mocks.ExecSafe != nil {
		return Mocks.ExecSafe(params)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: execSafe")
	defer span.Finish()

	if len(params) == 0 {
		return nil, nil, 0, errors.New("at least one argument required")
	}

	if !gitdomain.IsAllowedGitCmd(c.logger, params) {
		return nil, nil, 0, errors.Errorf("command failed: %q is not a allowed git command", params)
	}

	cmd := c.GitCommand(repo, params...)
	stdout, stderr, err = cmd.DividedOutput(ctx)
	exitCode = cmd.ExitStatus()
	if exitCode != 0 && err != nil {
		err = nil // the error must just indicate that the exit code was nonzero
	}
	return stdout, stderr, exitCode, err
}

// MergeBase returns the merge base commit for the specified commits.
func (c *ClientImplementor) MergeBase(ctx context.Context, repo api.RepoName, a, b api.CommitID) (api.CommitID, error) {
	if Mocks.MergeBase != nil {
		return Mocks.MergeBase(repo, a, b)
	}
	span, ctx := ot.StartSpanFromContext(ctx, "Git: MergeBase")
	span.SetTag("A", a)
	span.SetTag("B", b)
	defer span.Finish()

	cmd := c.GitCommand(repo, "merge-base", "--", string(a), string(b))
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
	}
	return api.CommitID(bytes.TrimSpace(out)), nil
}

// RevList makes a git rev-list call and iterates through the resulting commits, calling the provided onCommit function for each.
func (c *ClientImplementor) RevList(repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	command := c.GitCommand(api.RepoName(repo), RevListArgs(commit)...)
	command.DisableTimeout()
	stdout, err := command.StdoutReader(ctx)
	if err != nil {
		return err
	}
	defer stdout.Close()

	return c.RevListEach(stdout, onCommit)
}

func RevListArgs(givenCommit string) []string {
	return []string{"rev-list", "--first-parent", givenCommit}
}

func (c *ClientImplementor) RevListEach(stdout io.Reader, onCommit func(commit string) (shouldContinue bool, err error)) error {
	reader := bufio.NewReader(stdout)

	for {
		commit, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		commit = commit[:len(commit)-1] // Drop the trailing newline
		shouldContinue, err := onCommit(commit)
		if !shouldContinue {
			return err
		}
	}

	return nil
}

// GetBehindAhead returns the behind/ahead commit counts information for right vs. left (both Git
// revspecs).
func (c *ClientImplementor) GetBehindAhead(ctx context.Context, repo api.RepoName, left, right string) (*gitdomain.BehindAhead, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: BehindAhead")
	defer span.Finish()

	if err := checkSpecArgSafety(left); err != nil {
		return nil, err
	}
	if err := checkSpecArgSafety(right); err != nil {
		return nil, err
	}

	cmd := c.GitCommand(repo, "rev-list", "--count", "--left-right", fmt.Sprintf("%s...%s", left, right))
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
