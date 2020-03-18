package git

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

type Commit struct {
	ID        api.CommitID `json:"ID,omitempty"`
	Author    Signature    `json:"Author"`
	Committer *Signature   `json:"Committer,omitempty"`
	Message   string       `json:"Message,omitempty"`
	// Parents are the commit IDs of this commit's parent commits.
	Parents []api.CommitID `json:"Parents,omitempty"`
}

type Signature struct {
	Name  string    `json:"Name,omitempty"`
	Email string    `json:"Email,omitempty"`
	Date  time.Time `json:"Date"`
}

// CommitsOptions specifies options for (Repository).Commits (Repository).CommitCount.
type CommitsOptions struct {
	Range string // commit range (revspec, "A..B", "A...B", etc.)

	N    uint // limit the number of returned commits to this many (0 means no limit)
	Skip uint // skip this many commits at the beginning

	MessageQuery string // include only commits whose commit message contains this substring

	Author string // include only commits whose author matches this
	After  string // include only commits after this date

	Path string // only commits modifying the given path are selected (optional)

	// RemoteURLFunc is called to get the Git remote URL if it's not set in
	// repo and if it is needed. The Git remote URL is only required if the
	// gitserver doesn't already contain a clone of the repository or if the
	// commit must be fetched from the remote.
	RemoteURLFunc func() (string, error)
}

// logEntryPattern is the regexp pattern that matches entries in the output of the `git shortlog
// -sne` command.
var logEntryPattern = lazyregexp.New(`^\s*([0-9]+)\s+(.*)$`)

// getCommit returns the commit with the given id.
func getCommit(ctx context.Context, repo gitserver.Repo, remoteURLFunc func() (string, error), id api.CommitID) (*Commit, error) {
	if Mocks.GetCommit != nil {
		return Mocks.GetCommit(id)
	}

	if err := checkSpecArgSafety(string(id)); err != nil {
		return nil, err
	}

	commits, err := commitLog(ctx, repo, CommitsOptions{Range: string(id), N: 1, RemoteURLFunc: remoteURLFunc})
	if err != nil {
		return nil, err
	}

	if len(commits) != 1 {
		return nil, fmt.Errorf("git log: expected 1 commit, got %d", len(commits))
	}

	return commits[0], nil
}

// GetCommit returns the commit with the given commit ID, or ErrCommitNotFound if no such commit
// exists.
//
// The remoteURLFunc is called to get the Git remote URL if it's not set in repo and if it is
// needed. The Git remote URL is only required if the gitserver doesn't already contain a clone of
// the repository or if the commit must be fetched from the remote.
func GetCommit(ctx context.Context, repo gitserver.Repo, remoteURLFunc func() (string, error), id api.CommitID) (*Commit, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: GetCommit")
	span.SetTag("Commit", id)
	defer span.Finish()

	return getCommit(ctx, repo, remoteURLFunc, id)
}

// Commits returns all commits matching the options.
func Commits(ctx context.Context, repo gitserver.Repo, opt CommitsOptions) ([]*Commit, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Commits")
	span.SetTag("Opt", opt)
	defer span.Finish()

	if err := checkSpecArgSafety(string(opt.Range)); err != nil {
		return nil, err
	}

	return commitLog(ctx, repo, opt)
}

// HasCommitAfter indicates the staleness of a repository. It returns a boolean indicating if a repository
// contains a commit past a specified date.
func HasCommitAfter(ctx context.Context, repo gitserver.Repo, date string, revspec string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: HasCommitAfter")
	span.SetTag("Date", date)
	span.SetTag("RevSpec", revspec)
	defer span.Finish()

	if revspec == "" {
		revspec = "HEAD"
	}

	commitid, err := ResolveRevision(ctx, repo, nil, revspec, &ResolveRevisionOptions{NoEnsureRevision: true})
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

func isInvalidRevisionRangeError(output, obj string) bool {
	return strings.HasPrefix(output, "fatal: Invalid revision range "+obj)
}

// commitLog returns a list of commits.
//
// The caller is responsible for doing checkSpecArgSafety on opt.Head and opt.Base.
func commitLog(ctx context.Context, repo gitserver.Repo, opt CommitsOptions) (commits []*Commit, err error) {
	args, err := commitLogArgs([]string{"log", logFormatWithoutRefs}, opt)
	if err != nil {
		return nil, err
	}

	cmd := gitserver.DefaultClient.Command("git", args...)
	cmd.Repo = repo
	cmd.EnsureRevision = opt.Range
	retryer := &commandRetryer{
		cmd:           cmd,
		remoteURLFunc: opt.RemoteURLFunc,
		exec: func() error {
			commits, err = runCommitLog(ctx, cmd, opt)
			return err
		},
	}
	err = retryer.run()
	return
}

// runCommitLog sends the git command to gitserver. It interprets missing
// revision responses and converts them into RevisionNotFoundError.
func runCommitLog(ctx context.Context, cmd *gitserver.Cmd, opt CommitsOptions) ([]*Commit, error) {
	data, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		data = bytes.TrimSpace(data)
		if isBadObjectErr(string(stderr), string(opt.Range)) {
			return nil, &gitserver.RevisionNotFoundError{Repo: cmd.Repo.Name, Spec: string(opt.Range)}
		}
		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, data))
	}

	allParts := bytes.Split(data, []byte{'\x00'})
	numCommits := len(allParts) / partsPerCommit
	commits := make([]*Commit, 0, numCommits)
	for len(data) > 0 {
		var commit *Commit
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

// CommitCount returns the number of commits that would be returned by Commits.
func CommitCount(ctx context.Context, repo gitserver.Repo, opt CommitsOptions) (uint, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: CommitCount")
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
	out, err := cmd.CombinedOutput(ctx)
	if err != nil {
		return 0, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}
	out = bytes.TrimSpace(out)
	n, err := strconv.ParseUint(string(out), 10, 64)
	return uint(n), err
}

const (
	partsPerCommit = 10 // number of \x00-separated fields per commit

	// include refs (slow on repos with many refs)
	logFormatWithRefs = "--format=format:%H%x00%D%x00%aN%x00%aE%x00%at%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00"

	// don't include refs (faster, should be used if refs are not needed)
	logFormatWithoutRefs = "--format=format:%H%x00%x00%aN%x00%aE%x00%at%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00"
)

// parseCommitFromLog parses the next commit from data and returns the commit and the remaining
// data. The data arg is a byte array that contains NUL-separated log fields as formatted by
// logFormatFlag.
func parseCommitFromLog(data []byte) (commit *Commit, refs []string, rest []byte, err error) {
	parts := bytes.SplitN(data, []byte{'\x00'}, partsPerCommit+1)
	if len(parts) < partsPerCommit {
		return nil, nil, nil, fmt.Errorf("invalid commit log entry: %q", parts)
	}

	// log outputs are newline separated, so all but the 1st commit ID part
	// has an erroneous leading newline.
	parts[0] = bytes.TrimPrefix(parts[0], []byte{'\n'})
	commitID := api.CommitID(parts[0])

	authorTime, err := strconv.ParseInt(string(parts[4]), 10, 64)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parsing git commit author time: %s", err)
	}
	committerTime, err := strconv.ParseInt(string(parts[7]), 10, 64)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parsing git commit committer time: %s", err)
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

	commit = &Commit{
		ID:        commitID,
		Author:    Signature{Name: string(parts[2]), Email: string(parts[3]), Date: time.Unix(authorTime, 0).UTC()},
		Committer: &Signature{Name: string(parts[5]), Email: string(parts[6]), Date: time.Unix(committerTime, 0).UTC()},
		Message:   string(bytes.TrimSuffix(parts[8], []byte{'\n'})),
		Parents:   parents,
	}

	if len(parts) == partsPerCommit+1 {
		rest = parts[10]
	}

	return commit, refs, rest, nil
}

// onelineCommit contains (a subset of the) information about a commit returned
// by `git log --oneline --source`.
type onelineCommit struct {
	sha1      string // sha1 commit ID
	sourceRef string // `git log --source` source ref
}

// parseCommitsFromOnelineLog parses the commits from the output of:
//
//   git log --oneline -z --source --no-patch
func parseCommitsFromOnelineLog(data []byte) (commits []*onelineCommit, err error) {
	entries := bytes.Split(data, []byte{'\x00'})
	for _, e := range entries {
		if len(e) == 0 {
			continue
		}

		// Format: (40-char SHA) \t (source ref)? 'log size '
		if len(e) <= 40 {
			return commits, fmt.Errorf("parsing git oneline commit: short entry: %q", e)
		}
		sha1 := e[:40]
		i := bytes.Index(e, []byte{' '})
		if i == -1 {
			return commits, fmt.Errorf("parsing git oneline commit: no ' ': %q", e)
		}
		sourceRef := e[41:i]
		commits = append(commits, &onelineCommit{
			sha1:      string(sha1),
			sourceRef: string(sourceRef),
		})
	}
	return commits, nil
}
