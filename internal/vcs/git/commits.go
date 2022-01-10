package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// CommitsOptions specifies options for (Repository).Commits (Repository).CommitCount.
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

	// When true we opt out of attempting to fetch missing revisions
	NoEnsureRevision bool
}

// logEntryPattern is the regexp pattern that matches entries in the output of the `git shortlog
// -sne` command.
var logEntryPattern = lazyregexp.New(`^\s*([0-9]+)\s+(.*)$`)

var recordGetCommitQueries = os.Getenv("RECORD_GET_COMMIT_QUERIES") == "1"

// getCommit returns the commit with the given id.
func getCommit(ctx context.Context, repo api.RepoName, id api.CommitID, opt ResolveRevisionOptions) (_ *gitdomain.Commit, err error) {
	if Mocks.GetCommit != nil {
		return Mocks.GetCommit(id)
	}

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

	commits, err := commitLog(ctx, repo, commitOptions)
	if err != nil {
		return nil, err
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
func GetCommit(ctx context.Context, repo api.RepoName, id api.CommitID, opt ResolveRevisionOptions) (*gitdomain.Commit, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: GetCommit")
	span.SetTag("Commit", id)
	defer span.Finish()

	return getCommit(ctx, repo, id, opt)
}

// Commits returns all commits matching the options.
func Commits(ctx context.Context, repo api.RepoName, opt CommitsOptions) ([]*gitdomain.Commit, error) {
	if Mocks.Commits != nil {
		return Mocks.Commits(repo, opt)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: Commits")
	span.SetTag("Opt", opt)
	defer span.Finish()

	if err := checkSpecArgSafety(opt.Range); err != nil {
		return nil, err
	}

	return commitLog(ctx, repo, opt)
}

// CommitsUniqueToBranch returns a map from commits that exist on a particular
// branch in the given repository to their committer date. This set of commits is
// determined by listing `{branchName} ^HEAD`, which is interpreted as: all
// commits on {branchName} not also on the tip of the default branch. If the
// supplied branch name is the default branch, then this method instead returns
// all commits reachable from HEAD.
func CommitsUniqueToBranch(ctx context.Context, repo api.RepoName, branchName string, isDefaultBranch bool, maxAge *time.Time) (_ map[string]time.Time, err error) {
	args := []string{"log", "--pretty=format:%H:%cI"}
	if maxAge != nil {
		args = append(args, fmt.Sprintf("--after=%s", *maxAge))
	}
	if isDefaultBranch {
		args = append(args, "HEAD")
	} else {
		args = append(args, branchName, "^HEAD")
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return parseCommitsUniqueToBranch(strings.Split(string(out), "\n"))
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
func HasCommitAfter(ctx context.Context, repo api.RepoName, date string, revspec string) (bool, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: HasCommitAfter")
	span.SetTag("Date", date)
	span.SetTag("RevSpec", revspec)
	defer span.Finish()

	if revspec == "" {
		revspec = "HEAD"
	}

	commitid, err := ResolveRevision(ctx, repo, revspec, ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return false, err
	}

	n, err := CommitCount(ctx, repo, CommitsOptions{
		N:     1,
		After: date,
		Range: string(commitid),
	})
	return n > 0, err
}

func isBadObjectErr(output, obj string) bool {
	return output == "fatal: bad object "+obj
}

// commitLog returns a list of commits.
//
// The caller is responsible for doing checkSpecArgSafety on opt.Head and opt.Base.
func commitLog(ctx context.Context, repo api.RepoName, opt CommitsOptions) (commits []*gitdomain.Commit, err error) {
	args, err := commitLogArgs([]string{"log", logFormatWithoutRefs}, opt)
	if err != nil {
		return nil, err
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	if !opt.NoEnsureRevision {
		cmd.EnsureRevision = opt.Range
	}
	return runCommitLog(ctx, cmd, opt)
}

// runCommitLog sends the git command to gitserver. It interprets missing
// revision responses and converts them into RevisionNotFoundError.
// It is declared as a variable so that we can swap it out in tests
var runCommitLog = func(ctx context.Context, cmd *gitserver.Cmd, opt CommitsOptions) ([]*gitdomain.Commit, error) {
	data, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		data = bytes.TrimSpace(data)
		if isBadObjectErr(string(stderr), opt.Range) {
			return nil, &gitdomain.RevisionNotFoundError{Repo: cmd.Repo, Spec: opt.Range}
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, data))
	}

	allParts := bytes.Split(data, []byte{'\x00'})
	numCommits := len(allParts) / partsPerCommit
	commits := make([]*gitdomain.Commit, 0, numCommits)
	for len(data) > 0 {
		var commit *gitdomain.Commit
		var err error
		commit, _, data, err = parseCommitFromLog(data)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}
	return commits, nil
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

	if opt.Path != "" {
		args = append(args, "--", opt.Path)
	}
	return args, nil
}

// CommitCount returns the number of commits that would be returned by Commits.
func CommitCount(ctx context.Context, repo api.RepoName, opt CommitsOptions) (uint, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: CommitCount")
	span.SetTag("Opt", opt)
	defer span.Finish()

	args, err := commitLogArgs([]string{"rev-list", "--count"}, opt)
	if err != nil {
		return 0, err
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	if opt.Path != "" {
		// This doesn't include --follow flag because rev-list doesn't support it, so the number may be slightly off.
		cmd.Args = append(cmd.Args, "--", opt.Path)
	}
	out, err := cmd.Output(ctx)
	if err != nil {
		return 0, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}

	out = bytes.TrimSpace(out)
	n, err := strconv.ParseUint(string(out), 10, 64)
	return uint(n), err
}

// FirstEverCommit returns the first commit ever made to the repository.
func FirstEverCommit(ctx context.Context, repo api.RepoName) (*gitdomain.Commit, error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Git: FirstEverCommit")
	defer span.Finish()

	args := []string{"rev-list", "--max-count=1", "--max-parents=0", "HEAD"}
	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", args, out))
	}
	id := api.CommitID(bytes.TrimSpace(out))
	return GetCommit(ctx, repo, id, ResolveRevisionOptions{NoEnsureRevision: true})
}

// CommitExists determines if the given commit exists in the given repository.
func CommitExists(ctx context.Context, repo api.RepoName, id api.CommitID) (bool, error) {
	c, err := getCommit(ctx, repo, id, ResolveRevisionOptions{NoEnsureRevision: true})
	if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return c != nil, nil
}

// Head determines the tip commit of the default branch for the given repository.
// If no HEAD revision exists for the given repository (which occurs with empty
// repositories), a false-valued flag is returned along with a nil error and
// empty revision.
func Head(ctx context.Context, repo api.RepoName) (_ string, revisionExists bool, err error) {
	cmd := gitserver.DefaultClient.Command("git", "rev-parse", "HEAD")
	cmd.Repo = repo

	out, err := cmd.Output(ctx)
	if err != nil {
		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			err = nil
		}
		return "", false, err
	}

	return string(out), true, nil
}

const (
	partsPerCommit = 10 // number of \x00-separated fields per commit

	// don't include refs (faster, should be used if refs are not needed)
	logFormatWithoutRefs = "--format=format:%H%x00%x00%aN%x00%aE%x00%at%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00"
)

// parseCommitFromLog parses the next commit from data and returns the commit and the remaining
// data. The data arg is a byte array that contains NUL-separated log fields as formatted by
// logFormatFlag.
func parseCommitFromLog(data []byte) (commit *gitdomain.Commit, refs []string, rest []byte, err error) {
	parts := bytes.SplitN(data, []byte{'\x00'}, partsPerCommit+1)
	if len(parts) < partsPerCommit {
		return nil, nil, nil, errors.Errorf("invalid commit log entry: %q", parts)
	}

	// log outputs are newline separated, so all but the 1st commit ID part
	// has an erroneous leading newline.
	parts[0] = bytes.TrimPrefix(parts[0], []byte{'\n'})
	commitID := api.CommitID(parts[0])

	authorTime, err := strconv.ParseInt(string(parts[4]), 10, 64)
	if err != nil {
		return nil, nil, nil, errors.Errorf("parsing git commit author time: %s", err)
	}
	committerTime, err := strconv.ParseInt(string(parts[7]), 10, 64)
	if err != nil {
		return nil, nil, nil, errors.Errorf("parsing git commit committer time: %s", err)
	}

	var parents []api.CommitID
	if parentPart := parts[9]; len(parentPart) > 0 {
		parentIDs := bytes.Split(parentPart, []byte{' '})
		parents = make([]api.CommitID, len(parentIDs))
		for i, id := range parentIDs {
			parents[i] = api.CommitID(id)
		}
	}

	if len(parts[1]) > 0 {
		refs = strings.Split(string(parts[1]), ", ")
	}

	commit = &gitdomain.Commit{
		ID:        commitID,
		Author:    gitdomain.Signature{Name: string(parts[2]), Email: string(parts[3]), Date: time.Unix(authorTime, 0).UTC()},
		Committer: &gitdomain.Signature{Name: string(parts[5]), Email: string(parts[6]), Date: time.Unix(committerTime, 0).UTC()},
		Message:   gitdomain.Message(strings.TrimSuffix(string(parts[8]), "\n")),
		Parents:   parents,
	}

	if len(parts) == partsPerCommit+1 {
		rest = parts[10]
	}

	return commit, refs, rest, nil
}

// BranchesContaining returns a map from branch names to branch tip hashes for
// each branch containing the given commit.
func BranchesContaining(ctx context.Context, repo api.RepoName, commit api.CommitID) ([]string, error) {
	cmd := gitserver.DefaultClient.Command("git", "branch", "--contains", string(commit), "--format", "%(refname)")
	cmd.Repo = repo

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
func RefDescriptions(ctx context.Context, repo api.RepoName) (_ map[string][]gitdomain.RefDescription, err error) {
	args := []string{"for-each-ref", "--format=%(objectname):%(refname):%(HEAD):%(creatordate:iso8601-strict)"}
	for prefix := range refPrefixes {
		args = append(args, prefix)
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return parseRefDescriptions(strings.Split(string(out), "\n"))
}

var refPrefixes = map[string]gitdomain.RefType{
	"refs/heads/": gitdomain.RefTypeBranch,
	"refs/tags/":  gitdomain.RefTypeTag,
}

// parseRefDescriptions converts the output of the for-each-ref command in the RefDescriptions
// method to a map from commits to RefDescription objects. Each line should conform to the format
// string `%(objectname):%(refname):%(HEAD):%(creatordate)`, where
//
// - %(objectname) is the 40-character revhash
// - %(refname) is the name of the tag or branch (prefixed with refs/heads/ or ref/tags/)
// - %(HEAD) is `*` if the branch is the default branch (and whitesace otherwise)
// - %(creatordate) is the ISO-formatted date the object was created
func parseRefDescriptions(lines []string) (map[string][]gitdomain.RefDescription, error) {
	refDescriptions := make(map[string][]gitdomain.RefDescription, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 4)
		if len(parts) != 4 {
			return nil, errors.Errorf(`unexpected output from git for-each-ref "%s"`, line)
		}

		commit := parts[0]
		isDefaultBranch := parts[2] == "*"

		var name string
		var refType gitdomain.RefType
		for prefix, typ := range refPrefixes {
			if strings.HasPrefix(parts[1], prefix) {
				name = parts[1][len(prefix):]
				refType = typ
				break
			}
		}
		if refType == gitdomain.RefTypeUnknown {
			return nil, errors.Errorf(`unexpected output from git for-each-ref "%s"`, line)
		}

		createdDate, err := time.Parse(time.RFC3339, parts[3])
		if err != nil {
			return nil, errors.Errorf(`unexpected output from git for-each-ref (bad date format) "%s"`, line)
		}

		refDescriptions[commit] = append(refDescriptions[commit], gitdomain.RefDescription{
			Name:            name,
			Type:            refType,
			IsDefaultBranch: isDefaultBranch,
			CreatedDate:     createdDate,
		})
	}

	return refDescriptions, nil
}

// CommitDate returns the time that the given commit was committed. If the given
// revision does not exist, a false-valued flag is returned along with a nil
// error and zero-valued time.
func CommitDate(ctx context.Context, repo api.RepoName, commit api.CommitID) (_ string, _ time.Time, revisionExists bool, err error) {
	cmd := gitserver.DefaultClient.Command("git", "show", "-s", "--format=%H:%cI", string(commit))
	cmd.Repo = repo

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

type CommitGraphOptions struct {
	Commit  string
	AllRefs bool
	Limit   int
	Since   *time.Time
}

// CommitGraph returns the commit graph for the given repository as a mapping
// from a commit to its parents. If a commit is supplied, the returned graph will
// be rooted at the given commit. If a non-zero limit is supplied, at most that
// many commits will be returned.
func CommitGraph(ctx context.Context, repo api.RepoName, opts CommitGraphOptions) (_ *gitdomain.CommitGraph, err error) {
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

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo

	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return nil, err
	}

	return gitdomain.ParseCommitGraph(strings.Split(string(out), "\n")), nil
}
