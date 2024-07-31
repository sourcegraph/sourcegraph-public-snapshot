package rockskip

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/go-ctags"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	symbolparser "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// mockParser converts each line to a symbol.
type mockParser struct {
	parserType ctags_config.ParserType
}

func (p mockParser) Parse(path string, bytes []byte) ([]*ctags.Entry, error) {
	symbols := []*ctags.Entry{}

	for lineNumber, line := range strings.Split(string(bytes), "\n") {
		if line == "" {
			continue
		}

		symbols = append(symbols, &ctags.Entry{Name: line, Line: lineNumber + 1})
	}

	return symbols, nil
}

func (mockParser) Close() {}

func mockService(t *testing.T, git *subprocessGit) (*sql.DB, *Service) {
	observationCtx := &observation.TestContext
	db := dbtest.NewDB(t)
	repoFetcher := newMockRepositoryFetcher(git)
	createParser := func(parserType ctags_config.ParserType) (ctags.Parser, error) {
		return mockParser{parserType: parserType}, nil
	}

	// This ensures the ctags config is initialized properly
	conf.MockAndNotifyWatchers(&conf.Unified{})
	symbolParserPool, err := symbolparser.NewParserPool(observationCtx, "test", createParser, 1, symbolparser.DefaultParserTypes)
	require.NoError(t, err)
	service, err := NewService(observationCtx, db, git, repoFetcher, symbolParserPool, 1, 1, false, 1, 1, 1, false)
	require.NoError(t, err)

	return db, service
}

func gitRm(t *testing.T, repoDir string, state map[string][]string, filename string) {
	gitRun(t, repoDir, "rm", filename)
	delete(state, filename)
	gitRun(t, repoDir, "commit", "--allow-empty", "-m", "remove "+filename)
}

func gitAdd(t *testing.T, repoDir string, state map[string][]string, filename string, contents string) {
	require.NoError(t, os.WriteFile(path.Join(repoDir, filename), []byte(contents), 0644), "os.WriteFile")
	gitRun(t, repoDir, "add", filename)
	symbols, err := mockParser{ctags_config.UniversalCtags}.Parse(filename, []byte(contents))
	require.NoError(t, err)
	state[filename] = []string{}
	for _, symbol := range symbols {
		state[filename] = append(state[filename], symbol.Name)
	}
	gitRun(t, repoDir, "commit", "--allow-empty", "-m", "add "+filename)
}

func gitRun(t *testing.T, repoDir string, args ...string) {
	out, err := gitserver.CreateGitCommand(repoDir, "git", args...).CombinedOutput()
	require.NoError(t, err, string(out))
}

type subprocessGit struct {
	gitDir        string
	catFileCmd    *exec.Cmd
	catFileStdin  io.WriteCloser
	catFileStdout bufio.Reader
	gs            gitserver.Client
}

func newSubprocessGit(t testing.TB, gitDir string) (*subprocessGit, error) {
	cmd := exec.Command("git", "cat-file", "--batch")
	cmd.Dir = gitDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &subprocessGit{
		gitDir:        gitDir,
		catFileCmd:    cmd,
		catFileStdin:  stdin,
		catFileStdout: *bufio.NewReader(stdout),
		gs:            gitserver.NewTestClient(t),
	}, nil
}

func (g subprocessGit) Close() error {
	err := g.catFileStdin.Close()
	if err != nil {
		return err
	}
	return g.catFileCmd.Wait()
}

func (g *subprocessGit) ChangedFiles(ctx context.Context, repo api.RepoName, base string, head string) (gitserver.ChangedFilesIterator, error) {
	cmd := exec.CommandContext(ctx, "git",
		"diff-tree",
		"-r",
		"--root",
		"--format=format:",
		"--no-prefix",
		"--name-status",
		"--no-renames",
		"-z",
		head,
	)
	cmd.Dir = g.gitDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(byteutils.ScanNullLines)

	closeChan := make(chan struct{})
	closer := sync.OnceFunc(func() {
		_ = stdout.Close()
		_ = cmd.Wait()
		close(closeChan)
	})

	return &gitDiffIterator{
		rc:             stdout,
		scanner:        scanner,
		closeChan:      closeChan,
		onceFuncCloser: closer,
	}, nil
}

func (g *subprocessGit) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) (returnError error) {
	nextCursor := commit
	for {
		var commits []api.CommitID
		var err error
		commits, nextCursor, err = g.paginatedRevList(ctx, api.RepoName(repo), nextCursor, 100)
		if err != nil {
			return err
		}
		for _, c := range commits {
			shouldContinue, err := onCommit(string(c))
			if err != nil {
				return err
			}
			if !shouldContinue {
				return nil
			}
		}
		if nextCursor == "" {
			return nil
		}
	}
}

func (g *subprocessGit) paginatedRevList(ctx context.Context, repo api.RepoName, commit string, count int) (_ []api.CommitID, nextCursor string, _ error) {
	cmd := exec.Command("git", "log", logFormatWithoutRefs, "--first-parent", "-n", strconv.Itoa(count), commit)
	cmd.Dir = g.gitDir

	stdout, err := cmd.Output()
	if err != nil {
		return nil, "", err
	}

	commitScanner := bufio.NewScanner(bytes.NewReader(stdout))
	commitScanner.Split(commitSplitFunc)

	var commits []*gitdomain.Commit
	for commitScanner.Scan() {
		rawCommit := commitScanner.Bytes()
		commit, err := parseCommitFromLog(rawCommit)
		if err != nil {
			return nil, "", err
		}
		commits = append(commits, commit)
	}
	if err := commitScanner.Err(); err != nil {
		return nil, "", err
	}

	commitIDs := make([]api.CommitID, 0, count+1)

	for _, commit := range commits {
		commitIDs = append(commitIDs, commit.ID)
	}

	if len(commitIDs) > count {
		nextCursor = string(commitIDs[len(commitIDs)-1])
		commitIDs = commitIDs[:count]
	}

	return commitIDs, nextCursor, nil
}

func newMockRepositoryFetcher(git *subprocessGit) fetcher.RepositoryFetcher {
	return &mockRepositoryFetcher{git: git}
}

type mockRepositoryFetcher struct{ git *subprocessGit }

func (f *mockRepositoryFetcher) FetchRepositoryArchive(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) <-chan fetcher.ParseRequestOrError {
	ch := make(chan fetcher.ParseRequestOrError)

	go func() {
		for _, p := range paths {
			_, err := f.git.catFileStdin.Write([]byte(fmt.Sprintf("%s:%s\n", commit, p)))
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "writing to cat-file stdin"),
				}
				return
			}

			line, err := f.git.catFileStdout.ReadString('\n')
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "read newline"),
				}
				return
			}
			line = line[:len(line)-1] // Drop the trailing newline
			parts := strings.Split(line, " ")
			if len(parts) != 3 {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Newf("unexpected cat-file output: %q", line),
				}
				return
			}
			size, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "parse size"),
				}
				return
			}

			fileContents, err := io.ReadAll(io.LimitReader(&f.git.catFileStdout, size))
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "read contents"),
				}
				return
			}

			discarded, err := f.git.catFileStdout.Discard(1) // Discard the trailing newline
			if err != nil {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Wrap(err, "discard newline"),
				}
				return
			}
			if discarded != 1 {
				ch <- fetcher.ParseRequestOrError{
					Err: errors.Newf("expected to discard 1 byte, but discarded %d", discarded),
				}
				return
			}

			ch <- fetcher.ParseRequestOrError{
				ParseRequest: fetcher.ParseRequest{
					Path: p,
					Data: fileContents,
				},
			}
		}

		close(ch)
	}()

	return ch
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

// parseCommitFromLog parses the next commit from data and returns the commit and the remaining
// data. The data arg is a byte array that contains NUL-separated log fields as formatted by
// logFormatFlag.
func parseCommitFromLog(rawCommit []byte) (*gitdomain.Commit, error) {
	parts := bytes.Split(rawCommit, []byte{'\x00'})
	if len(parts) != partsPerCommit {
		return nil, errors.Newf("internal error: expected %d parts, got %d", partsPerCommit, len(parts))
	}

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

	return &gitdomain.Commit{
		ID:        commitID,
		Author:    gitdomain.Signature{Name: string(parts[1]), Email: string(parts[2]), Date: time.Unix(authorTime, 0).UTC()},
		Committer: &gitdomain.Signature{Name: string(parts[4]), Email: string(parts[5]), Date: time.Unix(committerTime, 0).UTC()},
		Message:   gitdomain.Message(strings.TrimSuffix(string(parts[7]), "\n")),
		Parents:   parents,
	}, nil
}

type gitDiffIterator struct {
	rc      io.ReadCloser
	scanner *bufio.Scanner

	closeChan      chan struct{}
	onceFuncCloser func()
}

func (i *gitDiffIterator) Next() (gitdomain.PathStatus, error) {
	select {
	case <-i.closeChan:
		return gitdomain.PathStatus{}, io.EOF
	default:
	}

	for i.scanner.Scan() {
		select {
		case <-i.closeChan:
			return gitdomain.PathStatus{}, io.EOF
		default:
		}

		status := i.scanner.Text()
		if len(status) == 0 {
			continue
		}

		if !i.scanner.Scan() {
			return gitdomain.PathStatus{}, errors.New("uneven pairs")
		}
		path := i.scanner.Text()

		switch status[0] {
		case 'A':
			return gitdomain.PathStatus{Path: path, Status: gitdomain.StatusAdded}, nil
		case 'M':
			return gitdomain.PathStatus{Path: path, Status: gitdomain.StatusModified}, nil
		case 'D':
			return gitdomain.PathStatus{Path: path, Status: gitdomain.StatusDeleted}, nil
		case 'T':
			return gitdomain.PathStatus{Path: path, Status: gitdomain.StatusTypeChanged}, nil
		default:
			return gitdomain.PathStatus{}, errors.Errorf("encountered unexpected file status %q for file %q", status, path)
		}
	}

	if err := i.scanner.Err(); err != nil {
		return gitdomain.PathStatus{}, errors.Wrap(err, "failed to scan git diff output")
	}

	return gitdomain.PathStatus{}, io.EOF
}

func (i *gitDiffIterator) Close() {
	i.onceFuncCloser()
}
