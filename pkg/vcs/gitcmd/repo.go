package gitcmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/util"
)

var (
	// gitCmdWhitelist are commands and arguments that are allowed to execute when calling GitCmdRaw.
	gitCmdWhitelist = map[string][]string{
		"log":    append([]string{}, gitCommonWhitelist...),
		"show":   append([]string{}, gitCommonWhitelist...),
		"remote": []string{"-v"},
		"diff":   append([]string{}, gitCommonWhitelist...),
		"blame":  []string{"--root", "--incremental", "--"},
		"branch": []string{"-r", "-a", "--contains"},

		"rev-parse":    []string{"--abbrev-ref", "--symbolic-full-name"},
		"rev-list":     []string{"--max-parents", "--reverse", "--max-count"},
		"ls-remote":    []string{"--get-url"},
		"symbolic-ref": []string{"--short"},
	}

	// `git log`, `git show`, `git diff`, etc., share a large common set of whitelisted args.
	gitCommonWhitelist = []string{
		"--name-status", "--full-history", "-M", "--date", "--format", "-i", "-n1", "-m", "--", "-n200", "-n2", "--follow", "--author", "--grep", "--date-order", "--decorate", "--skip", "--max-count", "--numstat", "--pretty", "--parents", "--topo-order", "--raw", "--follow", "--all", "--before", "--no-merges",
		"--patch", "--unified", "-S", "-G", "--pickaxe-all", "--pickaxe-regex", "--function-context", "--branches", "--source", "--src-prefix", "--dst-prefix", "--no-prefix",
		"--regexp-ignore-case", "--glob", "--cherry", "-z",
		"--until", "--since", "--author", "--committer",
		"--all-match", "--invert-grep", "--extended-regexp",
		"--no-color", "--decorate", "--no-patch", "--exclude",
		"--no-merges",
		"--full-index",
		"--find-copies",
		"--find-renames",
		"--inter-hunk-context",
	}

	reposDir = env.Get("SRC_REPOS_DIR", "", "Root dir containing repos.")
)

type Repository struct {
	// repoURI is the identifier of the repository, which is opaque to gitserver and is
	// conventionally a string like "github.com/gorilla/mux".
	repoURI api.RepoURI

	// remoteURL is the repository's Git remote URL.
	//
	// NOTE: Previously, the Git remote URL was derived from repoURI (by using the origin map). That
	// is still supported for backcompat, but it is now preferred to explicitly provide the Git
	// remote URL.
	remoteURL string
}

func (r *Repository) String() string {
	return fmt.Sprintf("git repo %s (remote: %s)", r.repoURI, r.remoteURL)
}

// Open returns a handle to a repository on gitserver with the given identifier (repoURI) and
// optional Git remote URL. The Git remote URL is only required if the gitserver doesn't already
// contain a clone of the repository or if EnsureRevision on a command is set to a revision that
// must be fetched from the remote.
func Open(repoURI api.RepoURI, remoteURL string) *Repository {
	return &Repository{repoURI: repoURI, remoteURL: remoteURL}
}

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which could
// cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.New("invalid git revision spec (begins with '-')")
	}
	return nil
}

// command creates a new gitserver.Cmd for the current repository. command
// name must be 'git', otherwise it panics.
func (r *Repository) command(name string, arg ...string) *gitserver.Cmd {
	cmd := gitserver.DefaultClient.Command(name, arg...)
	cmd.Repo = gitserver.Repo{Name: r.repoURI, URL: r.remoteURL}
	return cmd
}

// ResolveRevision will return the absolute commit for a commit-ish spec.
// If spec is empty, HEAD is used.
// Error cases:
// * Repo does not exist: vcs.RepoNotExistError
// * Commit does not exist: vcs.RevisionNotFoundError
// * Empty repository: vcs.RevisionNotFoundError
// * Other unexpected errors.
func (r *Repository) ResolveRevision(ctx context.Context, spec string, opt *vcs.ResolveRevisionOptions) (api.CommitID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ResolveRevision")
	span.SetTag("Spec", spec)
	span.SetTag("Opt", fmt.Sprintf("%+v", opt))
	defer span.Finish()

	if opt == nil {
		opt = &vcs.ResolveRevisionOptions{}
	}

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

	cmd := r.command("git", "rev-parse", spec)
	if !opt.NoEnsureRevision {
		cmd.EnsureRevision = string(spec)
	}
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return "", err
		}
		if bytes.Contains(stderr, []byte("unknown revision")) {
			return "", &vcs.RevisionNotFoundError{Repo: r.repoURI, Spec: spec}
		}
		return "", errors.WithMessage(err, fmt.Sprintf("exec `git rev-parse` failed with stderr: %s", stderr))
	}
	commit := api.CommitID(bytes.TrimSpace(stdout))
	if !vcs.IsAbsoluteRevision(string(commit)) {
		if commit == "HEAD" {
			// We don't verify the existence of HEAD (see above comments), but
			// if HEAD doesn't point to anything git just returns `HEAD` as the
			// output of rev-parse. An example where this occurs is an empty
			// repository.
			return "", &vcs.RevisionNotFoundError{Repo: r.repoURI, Spec: spec}
		}
		return "", fmt.Errorf("ResolveRevision: got bad commit %q for repo %q at revision %q", commit, r.repoURI, spec)
	}
	return commit, nil
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
		id := api.CommitID(ref[0])
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
			branch.Counts, err = r.BehindAhead(ctx, "refs/heads/"+opt.BehindAheadBranch, "refs/heads/"+name)
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
	cmd := r.command("git", append([]string{"branch"}, args...)...)
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

// BehindAhead returns the behind/ahead commit counts information for right vs. left (both Git revspecs).
func (r *Repository) BehindAhead(ctx context.Context, left, right string) (*vcs.BehindAhead, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: BehindAhead")
	defer span.Finish()

	if err := checkSpecArgSafety(left); err != nil {
		return nil, err
	}
	if err := checkSpecArgSafety(right); err != nil {
		return nil, err
	}

	cmd := r.command("git", "rev-list", "--count", "--left-right", fmt.Sprintf("%s...%s", left, right))
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

	cmd := r.command("git", "tag", "--list", "--sort", "-creatordate", "--format", "%(objectname)%00%(refname:short)%00%(creatordate:unix)")
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("listing git tags in %s failed: %s. Output was:\n\n%s", cmd.Repo, err, out)
	}

	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
	if len(out) == 0 {
		return nil, nil // no tags
	}
	lines := bytes.Split(out, []byte("\n"))
	tags := make([]*vcs.Tag, len(lines))
	for i, line := range lines {
		parts := bytes.SplitN(line, []byte("\x00"), 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid git tag list output line: %q", line)
		}
		date, err := strconv.ParseInt(string(parts[2]), 10, 64)
		if err != nil {
			return nil, err
		}
		tags[i] = &vcs.Tag{
			Name:        string(parts[1]),
			CommitID:    api.CommitID(parts[0]),
			CreatorDate: time.Unix(date, 0).UTC(),
		}
	}
	return tags, nil
}

type byteSlices [][]byte

func (p byteSlices) Len() int           { return len(p) }
func (p byteSlices) Less(i, j int) bool { return bytes.Compare(p[i], p[j]) < 0 }
func (p byteSlices) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (r *Repository) showRef(ctx context.Context, arg string) ([][2]string, error) {
	cmd := r.command("git", "show-ref", arg)
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

// getCommit returns the commit with the given id.
func (r *Repository) getCommit(ctx context.Context, id api.CommitID) (*vcs.Commit, error) {
	if err := checkSpecArgSafety(string(id)); err != nil {
		return nil, err
	}

	commits, err := r.commitLog(ctx, vcs.CommitsOptions{Range: string(id), N: 1})
	if err != nil {
		return nil, err
	}

	if len(commits) != 1 {
		return nil, fmt.Errorf("git log: expected 1 commit, got %d", len(commits))
	}

	return commits[0], nil
}

func (r *Repository) GetCommit(ctx context.Context, id api.CommitID) (*vcs.Commit, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: GetCommit")
	span.SetTag("Commit", id)
	defer span.Finish()

	return r.getCommit(ctx, id)
}

func (r *Repository) Commits(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Commits")
	span.SetTag("Opt", opt)
	defer span.Finish()

	if err := checkSpecArgSafety(string(opt.Range)); err != nil {
		return nil, err
	}

	return r.commitLog(ctx, opt)
}

func isBadObjectErr(output, obj string) bool {
	return string(output) == "fatal: bad object "+obj
}

func isInvalidRevisionRangeError(output, obj string) bool {
	return strings.HasPrefix(output, "fatal: Invalid revision range "+obj)
}

// commitLog returns a list of commits.
//
// The caller is responsible for doing checkSpecArgSafety on opt.Head and opt.Base.
func (r *Repository) commitLog(ctx context.Context, opt vcs.CommitsOptions) ([]*vcs.Commit, error) {
	args, err := commitLogArgs([]string{"log", logFormatWithoutRefs}, opt)
	if err != nil {
		return nil, err
	}

	cmd := r.command("git", args...)
	data, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		data = bytes.TrimSpace(data)
		if isBadObjectErr(string(stderr), string(opt.Range)) {
			return nil, &vcs.RevisionNotFoundError{Repo: r.repoURI, Spec: string(opt.Range)}
		}
		return nil, fmt.Errorf("exec `git log` failed: %s. Output was:\n\n%s", err, data)
	}

	allParts := bytes.Split(data, []byte{'\x00'})
	numCommits := len(allParts) / partsPerCommit
	commits := make([]*vcs.Commit, 0, numCommits)
	for len(data) > 0 {
		var commit *vcs.Commit
		var err error
		commit, _, data, err = parseCommitFromLog(data)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func commitLogArgs(initialArgs []string, opt vcs.CommitsOptions) (args []string, err error) {
	if err := checkSpecArgSafety(string(opt.Range)); err != nil {
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

	if opt.MessageQuery != "" {
		args = append(args, "--fixed-strings", "--regexp-ignore-case", "--grep="+opt.MessageQuery)
	}

	if opt.Range != "" {
		args = append(args, opt.Range)
	}

	if opt.Path != "" {
		args = append(args, "--", opt.Path)
	}
	return args, nil
}

func (r *Repository) CommitCount(ctx context.Context, opt vcs.CommitsOptions) (uint, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: CommitCount")
	span.SetTag("Opt", opt)
	defer span.Finish()

	args, err := commitLogArgs([]string{"rev-list", "--count"}, opt)
	if err != nil {
		return 0, err
	}

	cmd := r.command("git", args...)
	if opt.Path != "" {
		// This doesn't include --follow flag because rev-list doesn't support it, so the number may be slightly off.
		cmd.Args = append(cmd.Args, "--", opt.Path)
	}
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return 0, fmt.Errorf("exec `git rev-list --count` failed: %s. Output was:\n\n%s", err, out)
	}
	out = bytes.TrimSpace(out)
	return parseUint(string(out))
}

func parseUint(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	return uint(n), err
}

func (r *Repository) Diff(ctx context.Context, base, head api.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	r.ensureAbsCommit(base)
	r.ensureAbsCommit(head)
	if opt == nil {
		opt = &vcs.DiffOptions{}
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Diff")
	span.SetTag("Base", base)
	span.SetTag("Head", head)
	span.SetTag("Opt", opt)
	defer span.Finish()

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
	cmd := r.command("git", args...)
	if opt != nil {
		cmd.Args = append(cmd.Args, opt.Paths...)
	}
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		out = bytes.TrimSpace(out)
		if isBadObjectErr(string(out), string(base)) || isInvalidRevisionRangeError(string(out), string(base)) {
			return nil, &vcs.RevisionNotFoundError{Repo: r.repoURI, Spec: string(base)}
		}
		if isBadObjectErr(string(out), string(head)) || isInvalidRevisionRangeError(string(out), string(head)) {
			return nil, &vcs.RevisionNotFoundError{Repo: r.repoURI, Spec: string(head)}
		}
		return nil, fmt.Errorf("exec `git diff` failed: %s. Output was:\n\n%s", err, out)
	}
	diff := &vcs.Diff{Raw: string(out)}
	return diff, nil
}

// isWhitelistedGitArg checks if the arg is whitelisted.
func isWhitelistedGitArg(whitelistedArgs []string, arg string) bool {
	// Split the arg at the first equal sign and check the LHS against the whitelist args.
	splitArg := strings.Split(arg, "=")[0]
	for _, whiteListedArg := range whitelistedArgs {
		if splitArg == whiteListedArg {
			return true
		}
	}
	return false
}

// isWhitelistedGitCmd checks if the cmd and arguments are whitelisted.
func isWhitelistedGitCmd(args []string) bool {
	// check if the supplied command is a whitelisted cmd
	if len(gitCmdWhitelist) == 0 {
		return false
	}
	cmd := args[0]
	whiteListedArgs, ok := gitCmdWhitelist[cmd]
	if !ok {
		// Command not whitelisted
		return false
	}
	for _, arg := range args[1:] {
		if strings.HasPrefix(arg, "-") {
			// Special-case `git log -S` and `git log -G`, which interpret any characters
			// after their 'S' or 'G' as part of the query. There is no long form of this
			// flags (such as --something=query), so if we did not special-case these,
			// there would be no way to safely express a query that began with a '-'
			// character. (Same for `git show`, where the flag has the same meaning.)
			if (cmd == "log" || cmd == "show") && (strings.HasPrefix(arg, "-S") || strings.HasPrefix(arg, "-G")) {
				continue // this arg is OK
			}

			if !isWhitelistedGitArg(whiteListedArgs, arg) {
				return false
			}
		}
	}
	return true
}

func (r *Repository) GitCmdRaw(ctx context.Context, params []string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ExtensionGitCmd")
	defer span.Finish()

	if len(params) == 0 {
		return "", errors.New("at least one argument required")
	}

	if !isWhitelistedGitCmd(params) {
		return "", fmt.Errorf("command failed: %s is not a whitelisted git command", strings.Join(params, ""))
	}

	cmd := r.command("git", params...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", fmt.Errorf("exec git failed: %s. Command was:\n\n%s Output was:\n\n%s", err, strings.Join(params, ""), out)
	}

	return string(out), nil
}

func (r *Repository) ExecReader(ctx context.Context, args []string) (io.ReadCloser, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ExecReader")
	span.SetTag("args", args)
	defer span.Finish()

	if !isWhitelistedGitCmd(args) {
		return nil, fmt.Errorf("command failed: %v is not a whitelisted git command", args)
	}
	cmd := r.command("git", args...)
	return gitserver.StdoutReader(ctx, cmd)
}

func (r *Repository) BlameFileRaw(ctx context.Context, path string, opt *vcs.BlameOptions) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: BlameFile")
	span.SetTag(path, opt)
	defer span.Finish()

	if opt == nil {
		opt = &vcs.BlameOptions{}
	}
	if opt.OldestCommit != "" {
		return "", fmt.Errorf("OldestCommit not implemented")
	}
	if err := checkSpecArgSafety(string(opt.NewestCommit)); err != nil {
		return "", err
	}
	if err := checkSpecArgSafety(string(opt.OldestCommit)); err != nil {
		return "", err
	}

	args := []string{"blame", "--root", "--incremental"}
	if opt.StartLine != 0 || opt.EndLine != 0 {
		args = append(args, fmt.Sprintf("-L%d,%d", opt.StartLine, opt.EndLine))
	}
	args = append(args, string(opt.NewestCommit), "--", filepath.ToSlash(path))

	cmd := r.command("git", args...)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", fmt.Errorf("exec `git blame` failed: %s. Output was:\n\n%s", err, out)
	}
	if len(out) == 0 {
		return "", nil
	}

	return string(out[:]), nil
}

func (r *Repository) BlameFile(ctx context.Context, path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: BlameFile")
	span.SetTag(path, opt)
	defer span.Finish()

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

	cmd := r.command("git", args...)
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
			CommitID:  api.CommitID(commitID),
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
				ID:      api.CommitID(commitID),
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
			hunk.Message = commit.Message
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

func (r *Repository) MergeBase(ctx context.Context, a, b api.CommitID) (api.CommitID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: MergeBase")
	span.SetTag("A", a)
	span.SetTag("B", b)
	defer span.Finish()

	cmd := r.command("git", "merge-base", "--", string(a), string(b))
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return "", fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
	}
	return api.CommitID(bytes.TrimSpace(out)), nil
}

// logEntryPattern is the regexp pattern that matches entries in the output of the `git shortlog
// -sne` command.
var logEntryPattern = regexp.MustCompile(`^\s*([0-9]+)\s+(.*)$`)

func (r *Repository) ShortLog(ctx context.Context, opt vcs.ShortLogOptions) ([]*vcs.PersonCount, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ShortLog")
	span.SetTag("Opt", opt)
	defer span.Finish()

	if opt.Range == "" {
		opt.Range = "HEAD"
	}
	if err := checkSpecArgSafety(opt.Range); err != nil {
		return nil, err
	}

	args := []string{"shortlog", "-sne", "--no-merges"}
	if opt.After != "" {
		args = append(args, "--after="+opt.After)
	}
	args = append(args, opt.Range, "--")
	if opt.Path != "" {
		args = append(args, opt.Path)
	}
	cmd := r.command("git", args...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec `git shortlog -sne` failed: %v", err)
	}

	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, nil
	}
	lines := bytes.Split(out, []byte{'\n'})
	results := make([]*vcs.PersonCount, len(lines))
	for i, line := range lines {
		match := logEntryPattern.FindSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("invalid git shortlog line: %q", line)
		}
		count, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return nil, err
		}
		addr, err := mail.ParseAddress(string(match[2]))
		if err != nil || addr == nil {
			addr = &mail.Address{Name: string(match[2])}
		}
		results[i] = &vcs.PersonCount{
			Count: int32(count),
			Name:  addr.Name,
			Email: addr.Address,
		}
	}
	return results, nil
}

func (r *Repository) ReadFile(ctx context.Context, commit api.CommitID, name string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ReadFile")
	span.SetTag("Name", name)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	name = util.Rel(name)
	b, err := r.readFileBytes(ctx, commit, name)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *Repository) readFileBytes(ctx context.Context, commit api.CommitID, name string) ([]byte, error) {
	r.ensureAbsCommit(commit)

	cmd := r.command("git", "show", string(commit)+":"+name)
	cmd.EnsureRevision = string(commit)
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
	return out, nil
}

func (r *Repository) Lstat(ctx context.Context, commit api.CommitID, path string) (os.FileInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Lstat")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

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

func (r *Repository) Stat(ctx context.Context, commit api.CommitID, path string) (os.FileInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Stat")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	path = util.Rel(path)

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

func (r *Repository) ReadDir(ctx context.Context, commit api.CommitID, path string, recurse bool) ([]os.FileInfo, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ReadDir")
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
	return r.lsTree(ctx, commit, path, recurse)
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
func (r *Repository) lsTree(ctx context.Context, commit api.CommitID, path string, recurse bool) ([]os.FileInfo, error) {
	if path != "" || !recurse {
		// Only cache the root recursive ls-tree.
		return r.lsTreeUncached(ctx, commit, path, recurse)
	}

	key := string(r.repoURI) + ":" + string(commit) + ":" + path
	lsTreeRootCacheMu.Lock()
	v, ok := lsTreeRootCache.Get(key)
	lsTreeRootCacheMu.Unlock()
	var entries []os.FileInfo
	if ok {
		// Cache hit.
		entries = v.([]os.FileInfo)
	} else {
		// Cache miss.
		var err error
		start := time.Now()
		entries, err = r.lsTreeUncached(ctx, commit, path, recurse)
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

func (r *Repository) lsTreeUncached(ctx context.Context, commit api.CommitID, path string, recurse bool) ([]os.FileInfo, error) {
	r.ensureAbsCommit(commit)

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
	cmd := r.command("git", args...)
	cmd.EnsureRevision = string(commit)
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		if bytes.Contains(out, []byte("exists on disk, but not in")) {
			return nil, &os.PathError{Op: "ls-tree", Path: filepath.ToSlash(path), Err: os.ErrNotExist}
		}
		return nil, fmt.Errorf("exec %v failed: %s. Output was:\n\n%s", cmd.Args, err, out)
	}

	if len(out) == 0 {
		return nil, &os.PathError{Op: "git ls-tree", Path: path, Err: os.ErrNotExist}
	}

	trimPath := strings.TrimPrefix(path, "./")
	prefixLen := strings.LastIndexByte(trimPath, '/') + 1
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
		if len(name) < len(trimPath) {
			// This is in a submodule; return the original path to avoid a slice out of bounds panic
			// when setting the FileInfo._Name below.
			name = trimPath
		}

		if len(info) != 4 {
			return nil, fmt.Errorf("invalid `git ls-tree` output: %q", out)
		}
		typ := info[1]
		oid := info[2]
		if !vcs.IsAbsoluteRevision(oid) {
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
			cmd := r.command("git", "config", "--get", "submodule."+name+".url")
			url := "" // url is not available if submodules are not initialized
			if out, err := cmd.Output(ctx); err == nil {
				url = string(bytes.TrimSpace(out))
			}
			sys = vcs.SubmoduleInfo{
				URL:      url,
				CommitID: api.CommitID(oid),
			}
		case "tree":
			mode = mode | int64(os.ModeDir)
		}

		fis[i] = &util.FileInfo{
			// This returns the full relative path (e.g. "path/to/file.go") when the path arg is "./"
			// This behavior is necessary to construct the file tree.
			// In all other cases, it returns the basename (e.g. "file.go").
			Name_: name[prefixLen:],
			Mode_: os.FileMode(mode),
			Size_: size,
			Sys_:  sys,
		}
	}
	util.SortFileInfosByName(fis)

	return fis, nil
}

func (r *Repository) ensureAbsCommit(commitID api.CommitID) {
	// We don't want to even be running commands on non-absolute
	// commit IDs if we can avoid it, because we can't cache the
	// expensive part of those computations.
	if !vcs.IsAbsoluteRevision(string(commitID)) {
		panic(fmt.Errorf("non-absolute commit ID: %q on %s", commitID, r.String()))
	}
}

func (r *Repository) CreateCommitFromPatch(ctx context.Context, opt vcs.PatchOptions) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: CreatePhabricatorStagingObject")
	defer span.Finish()

	return gitserver.DefaultClient.CreateCommitFromPatch(ctx, r.repoURI, opt)
}
