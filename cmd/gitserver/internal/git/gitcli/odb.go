package gitcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) GetCommit(ctx context.Context, commit api.CommitID, includeModifiedFiles bool) (*git.GitCommitWithFiles, error) {
	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	args := buildGetCommitArgs(commit, includeModifiedFiles)

	r, err := g.NewCommand(ctx, WithArguments(args...))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	rawCommit, err := io.ReadAll(r)
	if err != nil {
		// If exit code is 128 and `fatal: bad object` is part of stderr, most likely we
		// are referencing a commit that does not exist.
		// We want to return a gitdomain.RevisionNotFoundError in that case.
		var e *CommandFailedError
		if errors.As(err, &e) && e.ExitStatus == 128 && bytes.Contains(e.Stderr, []byte("fatal: bad object")) {
			return nil, &gitdomain.RevisionNotFoundError{Repo: g.repoName, Spec: string(commit)}
		}

		return nil, err
	}

	c, err := parseCommitLogOutput(bytes.TrimPrefix(rawCommit, []byte{'\x1e'}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse commit log output")
	}
	return c, nil
}

func buildGetCommitArgs(commit api.CommitID, includeModifiedFiles bool) []string {
	args := []string{"log", logFormatWithoutRefs, "-n", "1"}
	if includeModifiedFiles {
		args = append(args, "--name-only")
	}
	args = append(args, string(commit))
	return args
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

func parseCommitLogOutput(rawCommit []byte) (*git.GitCommitWithFiles, error) {
	parts := bytes.Split(rawCommit, []byte{'\x00'})
	if len(parts) != partsPerCommit {
		return nil, errors.Newf("internal error: expected %d parts, got %d", partsPerCommit, len(parts))
	}

	return parseCommitFromLog(parts)
}

// parseCommitFromLog parses the next commit from data and returns the commit and the remaining
// data. The data arg is a byte array that contains NUL-separated log fields as formatted by
// logFormatFlag.
func parseCommitFromLog(parts [][]byte) (*git.GitCommitWithFiles, error) {
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

	var fileNames []string
	if fileOut := string(bytes.TrimSpace(parts[9])); fileOut != "" {
		fileNames = strings.Split(fileOut, "\n")
	}

	return &git.GitCommitWithFiles{
		Commit: &gitdomain.Commit{
			ID:        commitID,
			Author:    gitdomain.Signature{Name: string(parts[1]), Email: string(parts[2]), Date: time.Unix(authorTime, 0).UTC()},
			Committer: &gitdomain.Signature{Name: string(parts[4]), Email: string(parts[5]), Date: time.Unix(committerTime, 0).UTC()},
			Message:   gitdomain.Message(strings.TrimSuffix(string(parts[7]), "\n")),
			Parents:   parents,
		},
		ModifiedFiles: fileNames,
	}, nil
}

func (g *gitCLIBackend) ReadFile(ctx context.Context, commit api.CommitID, path string) (io.ReadCloser, error) {
	if err := gitdomain.EnsureAbsoluteCommit(commit); err != nil {
		return nil, err
	}

	blobOID, err := g.getBlobOID(ctx, commit, path)
	if err != nil {
		if err == errIsSubmodule {
			return io.NopCloser(bytes.NewReader(nil)), nil
		}
		return nil, err
	}

	return g.NewCommand(ctx, WithArguments("cat-file", "-p", string(blobOID)))
}

var errIsSubmodule = errors.New("blob is a submodule")

func (g *gitCLIBackend) getBlobOID(ctx context.Context, commit api.CommitID, path string) (api.CommitID, error) {
	out, err := g.NewCommand(ctx, WithArguments("ls-tree", string(commit), "--", path))
	if err != nil {
		return "", err
	}
	defer out.Close()

	stdout, err := io.ReadAll(out)
	if err != nil {
		// If exit code is 128 and `not a tree object` is part of stderr, most likely we
		// are referencing a commit that does not exist.
		// We want to return a gitdomain.RevisionNotFoundError in that case.
		var e *CommandFailedError
		if errors.As(err, &e) && e.ExitStatus == 128 {
			if bytes.Contains(e.Stderr, []byte("not a tree object")) || bytes.Contains(e.Stderr, []byte("Not a valid object name")) {
				return "", &gitdomain.RevisionNotFoundError{Repo: g.repoName, Spec: string(commit)}
			}
		}

		return "", err
	}

	stdout = bytes.TrimSpace(stdout)
	if len(stdout) == 0 {
		return "", &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}

	// format: 100644 blob 3bad331187e39c05c78a9b5e443689f78f4365a7	README.md
	fields := bytes.Fields(stdout)
	if len(fields) < 3 {
		return "", errors.Newf("unexpected output while parsing blob OID: %q", string(stdout))
	}
	if string(fields[0]) == "160000" {
		return "", errIsSubmodule
	}
	return api.CommitID(fields[2]), nil
}

func (g *gitCLIBackend) BehindAhead(ctx context.Context, left, right string) (*gitdomain.BehindAhead, error) {
	if err := checkSpecArgSafety(left); err != nil {
		return nil, err
	}
	if err := checkSpecArgSafety(right); err != nil {
		return nil, err
	}

	if left == "" {
		left = "HEAD"
	}

	if right == "" {
		right = "HEAD"
	}

	rc, err := g.NewCommand(ctx, WithArguments("rev-list", "--count", "--left-right", fmt.Sprintf("%s...%s", left, right)))
	if err != nil {
		return nil, errors.Wrap(err, "running git rev-list")
	}
	defer rc.Close()

	out, err := io.ReadAll(rc)
	if err != nil {
		var e *CommandFailedError
		if errors.As(err, &e) {
			switch {
			case e.ExitStatus == 128 && bytes.Contains(e.Stderr, []byte("fatal: ambiguous argument")):
				fallthrough
			case e.ExitStatus == 128 && bytes.Contains(e.Stderr, []byte("fatal: Invalid symmetric difference expression")):
				return nil, &gitdomain.RevisionNotFoundError{
					Repo: g.repoName,
					Spec: fmt.Sprintf("%s...%s", left, right),
				}
			}
		}

		return nil, errors.Wrap(err, "reading git rev-list output")
	}

	behindAhead := strings.Split(strings.TrimSuffix(string(out), "\n"), "\t")
	b, err := strconv.ParseUint(behindAhead[0], 10, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse behindahead output %q", out)
	}
	a, err := strconv.ParseUint(behindAhead[1], 10, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse behindahead output %q", out)
	}
	return &gitdomain.BehindAhead{Behind: uint32(b), Ahead: uint32(a)}, nil
}

func (g *gitCLIBackend) FirstEverCommit(ctx context.Context) (api.CommitID, error) {
	rc, err := g.NewCommand(ctx, WithArguments("rev-list", "--reverse", "--date-order", "--max-parents=0", "HEAD"))
	if err != nil {
		return "", err
	}
	defer rc.Close()

	out, err := io.ReadAll(rc)
	if err != nil {
		var cmdFailedErr *CommandFailedError
		if errors.As(err, &cmdFailedErr) {
			if cmdFailedErr.ExitStatus == 129 && bytes.Contains(cmdFailedErr.Stderr, []byte(revListUsageString)) {
				// If the error is due to an empty repository, return a sentinel error.
				e := &gitdomain.RevisionNotFoundError{
					Repo: g.repoName,
					Spec: "HEAD",
				}
				return "", e
			}
		}

		return "", errors.Wrap(err, "git rev-list command failed")
	}

	lines := bytes.TrimSpace(out)
	tokens := bytes.SplitN(lines, []byte("\n"), 2)
	if len(tokens) == 0 {
		return "", errors.New("FirstEverCommit returned no revisions")
	}
	first := tokens[0]
	id := api.CommitID(bytes.TrimSpace(first))
	return id, nil
}

const revListUsageString = `usage: git rev-list [<options>] <commit>... [--] [<path>...]`
