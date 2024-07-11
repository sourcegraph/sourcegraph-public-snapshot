package gitcli

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) CommitLog(ctx context.Context, opt git.CommitLogOpts) (git.CommitLogIterator, error) {
	for _, r := range opt.Ranges {
		if err := checkSpecArgSafety(r); err != nil {
			return nil, err
		}
	}

	if len(opt.Ranges) > 0 && opt.AllRefs {
		return nil, errors.New("cannot specify both a Range and AllRefs")
	}
	if len(opt.Ranges) == 0 && !opt.AllRefs {
		return nil, errors.New("must specify a Range or AllRefs")
	}

	args, err := buildCommitLogArgs(opt)
	if err != nil {
		return nil, err
	}

	r, err := g.NewCommand(ctx, WithArguments(args...))
	if err != nil {
		return nil, err
	}

	return newCommitLogIterator(g.repoName, strings.Join(opt.Ranges, " "), r), nil
}

func buildCommitLogArgs(opt git.CommitLogOpts) ([]string, error) {
	args := []string{"log", logFormatWithoutRefs}

	if opt.MaxCommits != 0 {
		args = append(args, "-n", strconv.FormatUint(uint64(opt.MaxCommits), 10))
	}
	if opt.Skip != 0 {
		args = append(args, "--skip="+strconv.FormatUint(uint64(opt.Skip), 10))
	}

	if opt.AuthorQuery != "" {
		args = append(args, "--fixed-strings", "--author="+opt.AuthorQuery)
	}

	if !opt.After.IsZero() {
		args = append(args, "--after="+opt.After.Format(time.RFC3339))
	}
	if !opt.Before.IsZero() {
		args = append(args, "--before="+opt.Before.Format(time.RFC3339))
	}
	switch opt.Order {
	case git.CommitLogOrderCommitDate:
		args = append(args, "--date-order")
	case git.CommitLogOrderTopoDate:
		args = append(args, "--topo-order")
	case git.CommitLogOrderDefault:
		// nothing to do
	default:
		return nil, errors.Newf("invalid ordering %d", opt.Order)
	}

	if opt.MessageQuery != "" {
		args = append(args, "--fixed-strings", "--regexp-ignore-case", "--grep="+opt.MessageQuery)
	}

	if opt.FollowOnlyFirstParent {
		args = append(args, "--first-parent")
	}

	if opt.AllRefs {
		args = append(args, "--all")
	}

	if opt.IncludeModifiedFiles {
		args = append(args, "--name-only")
	}
	if opt.FollowPathRenames {
		args = append(args, "--follow")
	}

	args = append(args, opt.Ranges...)

	args = append(args, "--")

	if opt.Path != "" {
		args = append(args, opt.Path)
	}

	return args, nil
}

func newCommitLogIterator(repoName api.RepoName, spec string, r io.ReadCloser) *commitLogIterator {
	commitScanner := bufio.NewScanner(r)
	// We use an increased buffer size since sub-repo permissions
	// can result in very lengthy output.
	commitScanner.Buffer(make([]byte, 0, 65536), 4294967296)
	commitScanner.Split(commitSplitFunc)

	return &commitLogIterator{
		Closer:   r,
		repoName: repoName,
		spec:     spec,
		sc:       commitScanner,
	}
}

type commitLogIterator struct {
	io.Closer
	repoName api.RepoName
	spec     string
	sc       *bufio.Scanner
}

func (it *commitLogIterator) Next() (*git.GitCommitWithFiles, error) {
	if !it.sc.Scan() {
		if err := it.sc.Err(); err != nil {
			// If exit code is 128 and `fatal: bad object` is part of stderr, most likely we
			// are referencing a commit that does not exist.
			// We want to return a gitdomain.RevisionNotFoundError in that case.
			var e *commandFailedError
			if errors.As(err, &e) && e.ExitStatus == 128 {
				if (bytes.Contains(e.Stderr, []byte("fatal: your current branch")) && bytes.Contains(e.Stderr, []byte("does not have any commits yet"))) || bytes.Contains(e.Stderr, []byte("fatal: bad revision 'HEAD'")) {
					return nil, io.EOF
				}

				// range with bad commit or bad ref on RHS: fatal: bad revision
				// range with bad commit or bad ref on LHS: fatal: Invalid revision range
				// 40 character commit sha: fatal: bad object
				// unknown ref name: fatal: ambiguous argument && unknown revision or path not in the working tree.

				var errMessages = []string{
					"not a tree object",
					"fatal: bad object",
					"fatal: Invalid revision range",
					"fatal: bad revision",
				}
				for _, message := range errMessages {
					if bytes.Contains(e.Stderr, []byte(message)) {
						return nil, &gitdomain.RevisionNotFoundError{Repo: it.repoName, Spec: it.spec}
					}
				}

				if bytes.Contains(e.Stderr, []byte("fatal: ambiguous argument")) && bytes.Contains(e.Stderr, []byte("unknown revision or path not in the working tree.")) {
					return nil, &gitdomain.RevisionNotFoundError{Repo: it.repoName, Spec: it.spec}
				}
			}
			return nil, err
		}
		return nil, io.EOF
	}

	rawCommit := it.sc.Bytes()
	commit, err := parseCommitFromLog(rawCommit)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (it *commitLogIterator) Close() error {
	err := it.Closer.Close()
	if err != nil {
		var e *commandFailedError
		if errors.As(err, &e) && e.ExitStatus == 128 {
			if (bytes.Contains(e.Stderr, []byte("fatal: your current branch")) && bytes.Contains(e.Stderr, []byte("does not have any commits yet"))) || bytes.Contains(e.Stderr, []byte("fatal: bad revision 'HEAD'")) {
				return nil
			}

			var errMessages = []string{
				"not a tree object",
				"fatal: bad object",
				"fatal: Invalid revision range",
				"fatal: bad revision",
			}
			for _, message := range errMessages {
				if bytes.Contains(e.Stderr, []byte(message)) {
					return &gitdomain.RevisionNotFoundError{Repo: it.repoName, Spec: it.spec}
				}
			}

			if bytes.Contains(e.Stderr, []byte("fatal: ambiguous argument")) && bytes.Contains(e.Stderr, []byte("unknown revision or path not in the working tree.")) {
				return &gitdomain.RevisionNotFoundError{Repo: it.repoName, Spec: it.spec}
			}
		}
	}
	return err
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
