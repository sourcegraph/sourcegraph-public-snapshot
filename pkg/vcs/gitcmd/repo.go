package gitcmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/binary"
	"github.com/golang/groupcache/lru"
	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cache"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

var (
	// logEntryPattern is the regexp pattern that matches entries in the output of
	// the `git shortlog -sne` command.
	logEntryPattern = regexp.MustCompile(`^\s*([0-9]+)\s+([A-Za-z]+(?:\s[A-Za-z]+)*)\s+<([A-Za-z@.]+)>\s*$`)
)

type Repository struct {
	URL string

	editLock sync.RWMutex // protects ops that change repository data
}

func (r *Repository) String() string {
	return fmt.Sprintf("git repo %s", r.URL)
}

func Open(url string) *Repository {
	return &Repository{URL: url}
}

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which could
// cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.New("invalid git revision spec (begins with '-')")
	}
	return nil
}

func (r *Repository) ResolveRevision(ctx context.Context, spec string) (vcs.CommitID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ResolveRevision")
	span.SetTag("Spec", spec)
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	if err := checkSpecArgSafety(spec); err != nil {
		return "", err
	}

	cmd := gitserver.DefaultClient.Command("git", "rev-parse", spec+"^0")
	cmd.Repo = r.URL
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return "", err
		}
		if bytes.Contains(stderr, []byte("unknown revision")) {
			return "", vcs.ErrRevisionNotFound
		}
		return "", fmt.Errorf("exec `git rev-parse` failed: %s. Stderr was:\n\n%s", err, stderr)
	}
	return vcs.CommitID(bytes.TrimSpace(stdout)), nil
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

func (r *Repository) Branches(ctx context.Context, opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Branches")
	span.SetTag("Opt", opt)
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	f := make(branchFilter)
	if opt.MergedInto != "" {
		b, err := r.branches(ctx, "--merged", opt.MergedInto)
		if err != nil {
			return nil, err
		}
		f.add(b)
	}
	if opt.ContainsCommit != "" {
		b, err := r.branches(ctx, "--contains="+opt.ContainsCommit)
		if err != nil {
			return nil, err
		}
		f.add(b)
	}

	refs, err := r.showRef(ctx, "--heads")
	if err != nil {
		return nil, err
	}

	var branches []*vcs.Branch
	for _, ref := range refs {
		name := strings.TrimPrefix(ref[1], "refs/heads/")
		id := vcs.CommitID(ref[0])
		if !f.allows(name) {
			continue
		}

		branch := &vcs.Branch{Name: name, Head: id}
		if opt.IncludeCommit {
			branch.Commit, err = r.getCommit(ctx, id)
			if err != nil {
				return nil, err
			}
		}
		if opt.BehindAheadBranch != "" {
			branch.Counts, err = r.branchesBehindAhead(ctx, name, opt.BehindAheadBranch)
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
func (r *Repository) branches(ctx context.Context, args ...string) ([]string, error) {
	cmd := gitserver.DefaultClient.Command("git", append([]string{"branch"}, args...)...)
	cmd.Repo = r.URL
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec %v in %s failed: %v (output follows)\n\n%s", cmd.Args, cmd.Repo, err, out)
	}
	lines := strings.Split(string(out), "\n")
	lines = lines[:len(lines)-1]
	branches := make([]string, len(lines))
	for i, line := range lines {
		branches[i] = line[2:]
	}
	return branches, nil
}

// branchesBehindAhead returns the behind/ahead commit counts information for branch, against base branch.
func (r *Repository) branchesBehindAhead(ctx context.Context, branch, base string) (*vcs.BehindAhead, error) {
	if err := checkSpecArgSafety(branch); err != nil {
		return nil, err
	}
	if err := checkSpecArgSafety(base); err != nil {
		return nil, err
	}

	cmd := gitserver.DefaultClient.Command("git", "rev-list", "--count", "--left-right", fmt.Sprintf("refs/heads/%s...refs/heads/%s", base, branch))
	cmd.Repo = r.URL
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
	return &vcs.BehindAhead{Behind: uint32(b), Ahead: uint32(a)}, nil
}

func (r *Repository) Tags(ctx context.Context) ([]*vcs.Tag, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Tags")
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	refs, err := r.showRef(ctx, "--tags")
	if err != nil {
		return nil, err
	}

	tags := make([]*vcs.Tag, len(refs))
	for i, ref := range refs {
		tags[i] = &vcs.Tag{
			Name:     strings.TrimPrefix(ref[1], "refs/tags/"),
			CommitID: vcs.CommitID(ref[0]),
		}
	}
	return tags, nil
}

type byteSlices [][]byte

func (p byteSlices) Len() int           { return len(p) }
func (p byteSlices) Less(i, j int) bool { return bytes.Compare(p[i], p[j]) < 0 }
func (p byteSlices) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (r *Repository) showRef(ctx context.Context, arg string) ([][2]string, error) {
	cmd := gitserver.DefaultClient.Command("git", "show-ref", arg)
	cmd.Repo = r.URL
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return nil, err
		}
		// Exit status of 1 and no output means there were no
		// results. This is not a fatal error.
		if cmd.ExitStatus == 1 && len(out) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("exec `git show-ref %s` in %s failed: %s. Output was:\n\n%s", arg, cmd.Repo, err, out)
	}

	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
	lines := bytes.Split(out, []byte("\n"))
	sort.Sort(byteSlices(lines)) // sort for consistency
	refs := make([][2]string, len(lines))
	for i, line := range lines {
		if len(line) <= 41 {
			return nil, errors.New("unexpectedly short (<=41 bytes) line in `git show-ref ...` output")
		}
		id := line[:40]
		name := line[41:]
		refs[i] = [2]string{string(id), string(name)}
	}
	return refs, nil
}

// getCommit returns the commit with the given id. The caller must be holding r.editLock.
func (r *Repository) getCommit(ctx context.Context, id vcs.CommitID) (*vcs.Commit, error) {
	if err := checkSpecArgSafety(string(id)); err != nil {
		return nil, err
	}

	commits, _, err := r.commitLog(ctx, vcs.CommitsOptions{Head: id, N: 1, NoTotal: true})
	if err != nil {
		return nil, err
	}

	if len(commits) != 1 {
		return nil, fmt.Errorf("git log: expected 1 commit, got %d", len(commits))
	}

	return commits[0], nil
}

func (r *Repository) GetCommit(ctx context.Context, id vcs.CommitID) (*vcs.Commit, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: GetCommit")
	span.SetTag("Commit", id)
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	return r.getCommit(ctx, id)
}

func (r *Repository) Commits(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Commits")
	span.SetTag("Opt", opt)
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	if err := checkSpecArgSafety(string(opt.Head)); err != nil {
		return nil, 0, err
	}
	if err := checkSpecArgSafety(string(opt.Base)); err != nil {
		return nil, 0, err
	}

	return r.commitLog(ctx, opt)
}

func isBadObjectErr(output, obj string) bool {
	return string(output) == "fatal: bad object "+obj
}

func isInvalidRevisionRangeError(output, obj string) bool {
	return strings.HasPrefix(output, "fatal: Invalid revision range "+obj)
}

var commitLogCache = cache.Sync(lru.New(500))

// commitLog returns a list of commits, and total number of commits
// starting from Head until Base or beginning of branch (unless NoTotal is true).
//
// The caller is responsible for doing checkSpecArgSafety on opt.Head and opt.Base.
func (r *Repository) commitLog(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	args := []string{"log", `--format=format:%H%x00%aN%x00%aE%x00%at%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00`}
	if opt.N != 0 {
		args = append(args, "-n", strconv.FormatUint(uint64(opt.N), 10))
	}
	if opt.Skip != 0 {
		args = append(args, "--skip="+strconv.FormatUint(uint64(opt.Skip), 10))
	}

	if opt.Path != "" {
		args = append(args, "--follow")
	}

	// Range
	rng := string(opt.Head)
	if opt.Base != "" {
		rng += "..." + string(opt.Base)
	}
	args = append(args, rng)

	if opt.Path != "" {
		args = append(args, "--", opt.Path)
	}

	// Only cache when we're fetching immutable data.
	var cacheKey string
	if len(opt.Head) == 40 && (len(opt.Base) == 0 || len(opt.Base) == 40) && opt.NoTotal {
		cacheKey = r.URL + "|" + fmt.Sprintf("%q", args)

		if commits, found := commitLogCache.Get(cacheKey); found {
			return commits.([]*vcs.Commit), 0, nil
		}
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = r.URL
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		out = bytes.TrimSpace(out)
		if isBadObjectErr(string(out), string(opt.Head)) {
			return nil, 0, vcs.ErrRevisionNotFound
		}
		return nil, 0, fmt.Errorf("exec `git log` failed: %s. Output was:\n\n%s", err, out)
	}

	const partsPerCommit = 9 // number of \x00-separated fields per commit
	allParts := bytes.Split(out, []byte{'\x00'})
	numCommits := len(allParts) / partsPerCommit
	commits := make([]*vcs.Commit, numCommits)
	for i := 0; i < numCommits; i++ {
		parts := allParts[partsPerCommit*i : partsPerCommit*(i+1)]

		// log outputs are newline separated, so all but the 1st commit ID part
		// has an erroneous leading newline.
		parts[0] = bytes.TrimPrefix(parts[0], []byte{'\n'})

		authorTime, err := strconv.ParseInt(string(parts[3]), 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing git commit author time: %s", err)
		}
		committerTime, err := strconv.ParseInt(string(parts[6]), 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing git commit committer time: %s", err)
		}

		var parents []vcs.CommitID
		if parentPart := parts[8]; len(parentPart) > 0 {
			parentIDs := bytes.Split(parentPart, []byte{' '})
			parents = make([]vcs.CommitID, len(parentIDs))
			for i, id := range parentIDs {
				parents[i] = vcs.CommitID(id)
			}
		}

		commits[i] = &vcs.Commit{
			ID:        vcs.CommitID(parts[0]),
			Author:    vcs.Signature{Name: string(parts[1]), Email: string(parts[2]), Date: time.Unix(authorTime, 0).UTC()},
			Committer: &vcs.Signature{Name: string(parts[4]), Email: string(parts[5]), Date: time.Unix(committerTime, 0).UTC()},
			Message:   string(bytes.TrimSuffix(parts[7], []byte{'\n'})),
			Parents:   parents,
		}
	}

	// Count commits.
	var total uint
	if !opt.NoTotal {
		cmd = gitserver.DefaultClient.Command("git", "rev-list", "--count", rng)
		if opt.Path != "" {
			// This doesn't include --follow flag because rev-list doesn't support it, so the number may be slightly off.
			cmd.Args = append(cmd.Args, "--", opt.Path)
		}
		cmd.Repo = r.URL
		out, err = cmd.CombinedOutput(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("exec `git rev-list --count` failed: %s. Output was:\n\n%s", err, out)
		}
		out = bytes.TrimSpace(out)
		total, err = parseUint(string(out))
		if err != nil {
			return nil, 0, err
		}
	}

	if cacheKey != "" {
		commitLogCache.Add(cacheKey, commits)
	}

	return commits, total, nil
}

func parseUint(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	return uint(n), err
}

var diffCache = cache.Sync(lru.New(100))

func (r *Repository) Diff(ctx context.Context, base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	ensureAbsCommit(base)
	ensureAbsCommit(head)
	if opt == nil {
		opt = &vcs.DiffOptions{}
	}
	optData, err := binary.Marshal(opt)
	if err != nil {
		return nil, err
	}
	cacheKey := r.URL + "|" + string(base) + "|" + string(head) + "|" + string(optData)
	if diff, found := diffCache.Get(cacheKey); found {
		return diff.(*vcs.Diff), nil
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Diff")
	span.SetTag("Base", base)
	span.SetTag("Head", head)
	span.SetTag("Opt", opt)
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	if strings.HasPrefix(string(base), "-") || strings.HasPrefix(string(head), "-") {
		// Protect against base or head that is interpreted as command-line option.
		return nil, errors.New("diff revspecs must not start with '-'")
	}

	if opt == nil {
		opt = &vcs.DiffOptions{}
	}
	args := []string{"diff", "--full-index"}
	if opt.DetectRenames {
		args = append(args, "-M")
	}
	args = append(args, "--src-prefix="+opt.OrigPrefix)
	args = append(args, "--dst-prefix="+opt.NewPrefix)

	rng := string(base)
	if opt.ExcludeReachableFromBoth {
		rng += "..." + string(head)
	} else {
		rng += ".." + string(head)
	}

	args = append(args, rng, "--")
	cmd := gitserver.DefaultClient.Command("git", args...)
	if opt != nil {
		cmd.Args = append(cmd.Args, opt.Paths...)
	}
	cmd.Repo = r.URL
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		out = bytes.TrimSpace(out)
		if isBadObjectErr(string(out), string(base)) || isBadObjectErr(string(out), string(head)) || isInvalidRevisionRangeError(string(out), string(base)) || isInvalidRevisionRangeError(string(out), string(head)) {
			return nil, vcs.ErrRevisionNotFound
		}
		return nil, fmt.Errorf("exec `git diff` failed: %s. Output was:\n\n%s", err, out)
	}
	diff := &vcs.Diff{Raw: string(out)}
	diffCache.Add(cacheKey, diff)
	return diff, nil
}

// UpdateEverything updates all branches, tags, etc., to match the
// default remote repository.
func (r *Repository) UpdateEverything(ctx context.Context, opt vcs.RemoteOpts) (*vcs.UpdateResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: UpdateEverything")
	span.SetTag("Opt", opt)
	defer span.Finish()

	r.editLock.Lock()
	defer r.editLock.Unlock()

	cmd := gitserver.DefaultClient.Command("git", "remote", "update", "--prune")
	cmd.Repo = r.URL
	cmd.Opt = &opt
	_, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec `git remote update` failed: %v. Stderr was:\n\n%s", err, string(stderr))
	}
	result, err := parseRemoteUpdate(stderr)
	if err != nil {
		return nil, fmt.Errorf("parsing output of `git remote update` failed: %v", err)
	}
	return &result, nil
}

var blameCache = cache.Sync(lru.New(500))

func (r *Repository) BlameFile(ctx context.Context, path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: BlameFile")
	span.SetTag(path, opt)
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	if opt == nil {
		opt = &vcs.BlameOptions{}
	}
	if opt.OldestCommit != "" {
		return nil, fmt.Errorf("OldestCommit not implemented")
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

	// Only cache when we're fetching immutable data (head is an
	// absolute commit ID, base is empty or an absolute commit
	// ID). Also, it's probably not worth caching blames of mere
	// regions of a file.
	var cacheKey string
	if len(opt.NewestCommit) == 40 && (len(opt.OldestCommit) == 0 || len(opt.OldestCommit) == 40) && opt.StartLine == 0 && opt.EndLine == 0 {
		cacheKey = r.URL + "|" + fmt.Sprintf("%q", args)

		if hunks, found := blameCache.Get(cacheKey); found {
			return hunks.([]*vcs.Hunk), nil
		}
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = r.URL
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec `git blame` failed: %s. Output was:\n\n%s", err, out)
	}
	if len(out) == 0 {
		return nil, nil
	}

	commits := make(map[string]vcs.Commit)
	hunks := make([]*vcs.Hunk, 0)
	remainingLines := strings.Split(string(out[:len(out)-1]), "\n")
	byteOffset := 0
	for len(remainingLines) > 0 {
		// Consume hunk
		hunkHeader := strings.Split(remainingLines[0], " ")
		if len(hunkHeader) != 4 {
			return nil, fmt.Errorf("Expected at least 4 parts to hunkHeader, but got: '%s'", hunkHeader)
		}
		commitID := hunkHeader[0]
		lineNoCur, _ := strconv.Atoi(hunkHeader[2])
		nLines, _ := strconv.Atoi(hunkHeader[3])
		hunk := &vcs.Hunk{
			CommitID:  vcs.CommitID(commitID),
			StartLine: int(lineNoCur),
			EndLine:   int(lineNoCur + nLines),
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
				return nil, fmt.Errorf("Failed to parse author-time %q", remainingLines[3])
			}
			summary := strings.Join(strings.Split(remainingLines[9], " ")[1:], " ")
			commit := vcs.Commit{
				ID:      vcs.CommitID(commitID),
				Message: summary,
				Author: vcs.Signature{
					Name:  author,
					Email: email,
					Date:  time.Unix(authorTime, 0).UTC(),
				},
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
				return nil, fmt.Errorf("Unexpected number of remaining lines (%d):\n%s", len(remainingLines), "  "+strings.Join(remainingLines, "\n  "))
			}

			commits[commitID] = commit
		}

		if commit, present := commits[commitID]; present {
			// Should always be present, but check just to avoid
			// panicking in case of a (somewhat likely) bug in our
			// git-blame parser above.
			hunk.CommitID = commit.ID
			hunk.Author = commit.Author
		}

		// Consume remaining lines in hunk
		for i := 1; i < nLines; i++ {
			byteOffset += len(remainingLines[1])
			remainingLines = remainingLines[2:]
		}

		hunk.EndByte = byteOffset
		hunks = append(hunks, hunk)
	}

	if cacheKey != "" {
		blameCache.Add(cacheKey, hunks)
	}

	return hunks, nil
}

func (r *Repository) MergeBase(ctx context.Context, a, b vcs.CommitID) (vcs.CommitID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: MergeBase")
	span.SetTag("A", a)
	span.SetTag("B", b)
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	cmd := gitserver.DefaultClient.Command("git", "merge-base", "--", string(a), string(b))
	cmd.Repo = r.URL
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
	}
	return vcs.CommitID(bytes.TrimSpace(out)), nil
}

func (r *Repository) Committers(ctx context.Context, opt vcs.CommittersOptions) ([]*vcs.Committer, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Committers")
	span.SetTag("Opt", opt)
	defer span.Finish()

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	if opt.Rev == "" {
		opt.Rev = "HEAD"
	}

	cmd := gitserver.DefaultClient.Command("git", "shortlog", "-sne", opt.Rev)
	cmd.Repo = r.URL
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec `git shortlog -sne` failed: %v", err)
	}
	out = bytes.TrimSpace(out)

	allEntries := bytes.Split(out, []byte{'\n'})
	numEntries := len(allEntries)
	if opt.N > 0 && numEntries > opt.N {
		numEntries = opt.N
	}
	var committers []*vcs.Committer
	for i := 0; i < numEntries; i++ {
		line := string(allEntries[i])
		if match := logEntryPattern.FindStringSubmatch(line); match != nil {
			commits, err2 := strconv.Atoi(match[1])
			if err2 != nil {
				continue
			}
			committers = append(committers, &vcs.Committer{
				Commits: int32(commits),
				Name:    match[2],
				Email:   match[3],
			})
		}
	}
	return committers, nil
}

func (r *Repository) ReadFile(ctx context.Context, commit vcs.CommitID, name string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ReadFile")
	span.SetTag("Name", name)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	name = util.Rel(name)
	r.editLock.RLock()
	defer r.editLock.RUnlock()
	b, err := r.readFileBytes(ctx, commit, name)
	if err != nil {
		return nil, err
	}
	return b, nil
}

var readFileBytesCache = cache.Sync(lru.New(1000))

func (r *Repository) readFileBytes(ctx context.Context, commit vcs.CommitID, name string) ([]byte, error) {
	ensureAbsCommit(commit)
	cacheKey := r.URL + "|" + string(commit) + "|" + name
	if data, found := readFileBytesCache.Get(cacheKey); found {
		return data.([]byte), nil
	}

	cmd := gitserver.DefaultClient.Command("git", "show", string(commit)+":"+name)
	cmd.Repo = r.URL
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contains(out, []byte("exists on disk, but not in")) || bytes.Contains(out, []byte("does not exist")) {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		if bytes.HasPrefix(out, []byte("fatal: bad object ")) {
			// Could be a git submodule.
			fi, err := r.Stat(ctx, commit, name)
			if err != nil {
				return nil, err
			}
			// Return empty for a submodule for now.
			if fi.Mode()&vcs.ModeSubmodule != 0 {
				return nil, nil
			}

		}
		return nil, fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
	}
	readFileBytesCache.Add(cacheKey, out)
	return out, nil
}

func (r *Repository) Lstat(ctx context.Context, commit vcs.CommitID, path string) (os.FileInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Lstat")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	path = filepath.Clean(util.Rel(path))

	if path == "." {
		// Special case root, which is not returned by `git ls-tree`.
		return &util.FileInfo{Mode_: os.ModeDir}, nil
	}

	fis, err := r.lsTree(ctx, commit, path, false)
	if err != nil {
		return nil, err
	}
	if len(fis) == 0 {
		return nil, &os.PathError{Op: "ls-tree", Path: path, Err: os.ErrNotExist}
	}

	return fis[0], nil
}

func (r *Repository) Stat(ctx context.Context, commit vcs.CommitID, path string) (os.FileInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Stat")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	path = util.Rel(path)

	r.editLock.RLock()
	defer r.editLock.RUnlock()

	fi, err := r.Lstat(ctx, commit, path)
	if err != nil {
		return nil, err
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		// Deref symlink.
		b, err := r.readFileBytes(ctx, commit, path)
		if err != nil {
			return nil, err
		}
		fi2, err := r.Lstat(ctx, commit, string(b))
		if err != nil {
			return nil, err
		}
		fi2.(*util.FileInfo).Name_ = fi.Name()
		return fi2, nil
	}

	return fi, nil
}

func (r *Repository) ReadDir(ctx context.Context, commit vcs.CommitID, path string, recurse bool) ([]os.FileInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ReadDir")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	span.SetTag("Recurse", recurse)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	r.editLock.RLock()
	defer r.editLock.RUnlock()
	// Trailing slash is necessary to ls-tree under the dir (not just
	// to list the dir's tree entry in its parent dir).
	return r.lsTree(ctx, commit, filepath.Clean(util.Rel(path))+"/", recurse)
}

var lsTreeCache = cache.Sync(lru.New(10000))

// lsTree returns ls of tree at path. The caller must be holding r.editLock.RLock().
func (r *Repository) lsTree(ctx context.Context, commit vcs.CommitID, path string, recurse bool) ([]os.FileInfo, error) {
	ensureAbsCommit(commit)
	cacheKey := r.URL + "|" + string(commit) + "|" + path + "|" + strconv.FormatBool(recurse)
	if fis, found := lsTreeCache.Get(cacheKey); found {
		return fis.([]os.FileInfo), nil
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
	args = append(args, "--", filepath.ToSlash(path))
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = r.URL
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contains(out, []byte("exists on disk, but not in")) {
			return nil, &os.PathError{Op: "ls-tree", Path: filepath.ToSlash(path), Err: os.ErrNotExist}
		}
		return nil, fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
	}

	if len(out) == 0 {
		return nil, os.ErrNotExist
	}

	prefixLen := strings.LastIndexByte(strings.TrimPrefix(path, "./"), '/') + 1
	lines := strings.Split(string(out), "\x00")
	fis := make([]os.FileInfo, len(lines)-1)
	for i, line := range lines {
		if i == len(lines)-1 {
			// last entry is empty
			continue
		}

		tabPos := strings.IndexByte(line, '\t')
		if tabPos == -1 {
			return nil, fmt.Errorf("invalid `git ls-tree` output: %q", out)
		}
		info := strings.SplitN(line[:tabPos], " ", 4)
		name := line[tabPos+1:]

		if len(info) != 4 {
			return nil, fmt.Errorf("invalid `git ls-tree` output: %q", out)
		}
		typ := info[1]
		oid := info[2]
		if len(oid) != 40 {
			return nil, fmt.Errorf("invalid `git ls-tree` oid output: %q", oid)
		}

		sizeStr := strings.TrimSpace(info[3])
		var size int64
		if sizeStr != "-" {
			// Size of "-" indicates a dir or submodule.
			size, err = strconv.ParseInt(sizeStr, 10, 64)
			if err != nil || size < 0 {
				return nil, fmt.Errorf("invalid `git ls-tree` size output: %q (error: %s)", sizeStr, err)
			}
		}

		var sys interface{}
		mode, err := strconv.ParseInt(info[0], 8, 32)
		if err != nil {
			return nil, err
		}
		switch typ {
		case "blob":
			const gitModeSymlink = 020000
			if mode&gitModeSymlink != 0 {
				mode = int64(os.ModeSymlink)
			} else {
				// Regular file.
				mode = mode | 0644
			}
		case "commit":
			mode = mode | vcs.ModeSubmodule
			cmd := gitserver.DefaultClient.Command("git", "config", "--get", "submodule."+name+".url")
			cmd.Repo = r.URL
			url := "" // url is not available if submodules are not initialized
			if out, err := cmd.Output(ctx); err == nil {
				url = string(bytes.TrimSpace(out))
			}
			sys = vcs.SubmoduleInfo{
				URL:      url,
				CommitID: vcs.CommitID(oid),
			}
		case "tree":
			mode = mode | int64(os.ModeDir)
		}

		fis[i] = &util.FileInfo{
			Name_: name[prefixLen:],
			Mode_: os.FileMode(mode),
			Size_: size,
			Sys_:  sys,
		}
	}
	util.SortFileInfosByName(fis)

	lsTreeCache.Add(cacheKey, fis)
	return fis, nil
}

func ensureAbsCommit(commitID vcs.CommitID) {
	// We don't want to even be running commands on non-absolute
	// commit IDs if we can avoid it, because we can't cache the
	// expensive part of those computations.
	if len(commitID) != 40 {
		panic(fmt.Errorf("non-absolute commit ID: %q", commitID))
	}
}
