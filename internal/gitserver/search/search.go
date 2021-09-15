package search

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

// TODO(camdencheek): this is copied straight from internal/search to avoid import cycles
type RevisionSpecifier struct {
	// RevSpec is a revision range specifier suitable for passing to git. See
	// the manpage gitrevisions(7).
	RevSpec string

	// RefGlob is a reference glob to pass to git. See the documentation for
	// "--glob" in git-log.
	RefGlob string

	// ExcludeRefGlob is a glob for references to exclude. See the
	// documentation for "--exclude" in git-log.
	ExcludeRefGlob string
}

// Git formatting directives as described in man git-log / PRETTY FORMATS
const (
	hash           = "%H"
	refNames       = "%D"
	authorName     = "%aN"
	authorEmail    = "%aE"
	authorDate     = "%at"
	committerName  = "%cN"
	committerEmail = "%cE"
	committerDate  = "%ct"
	rawBody        = "%B"
	parentHashes   = "%P"
)

var (
	partsPerCommit = len(formatWithRefs)

	formatWithRefs = []string{
		hash,
		refNames,
		authorName,
		authorEmail,
		authorDate,
		committerName,
		committerEmail,
		committerDate,
		rawBody,
		parentHashes,
	}

	formatWithoutRefs = []string{
		hash,
		"",
		authorName,
		authorEmail,
		authorDate,
		committerName,
		committerEmail,
		committerDate,
		rawBody,
		parentHashes,
	}

	logArgsWithRefs = []string{
		"log",
		"--decorate=full",
		"-z",
		"--no-merges",
		"--format=format:%s" + strings.Join(formatWithRefs, "%x00") + "%x00",
	}

	logArgsWithoutRefs = []string{
		"log",
		"--decorate=full",
		"-z",
		"--no-merges",
		"--format=format:%s" + strings.Join(formatWithoutRefs, "%x00") + "%x00",
	}

	sep = []byte{0x0}
)

type CommitView struct {
	Hash           []byte
	RefNames       []byte
	AuthorName     []byte
	AuthorEmail    []byte
	AuthorDate     []byte
	CommitterName  []byte
	CommitterEmail []byte
	CommitterDate  []byte
	Message        []byte
	ParentHashes   []byte
}

func NewCommitScanner(r io.Reader) *CommitScanner {
	scanner := bufio.NewScanner(r)

	// Split by commit
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if len(data) == 0 { // should only happen when atEOF
			return 0, nil, nil
		}

		// See if we have enough null bytes to constitute a full commit
		if bytes.Count(data, sep) < partsPerCommit+1 {
			if atEOF {
				return 0, nil, errors.Errorf("incomplete line")
			}
			return 0, nil, nil
		}

		// If we do, expand token to the end of that commit
		for i := 0; i < partsPerCommit+1; i++ {
			idx := bytes.IndexByte(data[len(token):], 0x0)
			if idx == -1 {
				panic("we already counted enough bytes in data")
			}
			token = data[:len(token)+idx+1]
		}
		return len(token), token, nil

	})

	return &CommitScanner{
		scanner: scanner,
	}
}

type CommitScanner struct {
	scanner *bufio.Scanner
	next    CommitView
	err     error
}

func (c *CommitScanner) Scan() bool {
	if !c.scanner.Scan() {
		return false
	}

	parts := bytes.SplitN(c.scanner.Bytes(), sep, partsPerCommit+1)
	if len(parts) < partsPerCommit+1 {
		c.err = errors.Errorf("invalid commit log entry: %q", parts)
		return false
	}

	c.next.Hash = parts[0]
	c.next.RefNames = parts[1]
	c.next.AuthorName = parts[2]
	c.next.AuthorEmail = parts[3]
	c.next.AuthorDate = parts[4]
	c.next.CommitterName = parts[5]
	c.next.CommitterEmail = parts[6]
	c.next.CommitterDate = parts[7]
	c.next.Message = parts[8]
	c.next.ParentHashes = parts[9]

	return true
}

// CommitView returns a parsed view of the formatted commit as output by git.
// The view is only valid until CommitScanner.Scan() is called next, and must
// be copied if you want its lifetime to exceed that.
func (c *CommitScanner) NextCommitView() *CommitView {
	return &c.next
}

func (c *CommitScanner) Err() error {
	return c.err
}

func Search(dir string, revs []RevisionSpecifier, p CommitPredicate, onMatch func(*LazyCommit, *HighlightedCommit) bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	diffFetcher, err := StartDiffFetcher(ctx, dir)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "git", logArgsWithoutRefs...)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		return err
	}

	var cmdErr error
	go func() {
		cmdErr = cmd.Wait()
		pw.Close()
	}()

	cs := NewCommitScanner(pr)
	for cs.Scan() {
		cv := cs.NextCommitView()
		lc := &LazyCommit{
			CommitView:  cv,
			diffFetcher: diffFetcher,
		}
		commitMatches, highlights, err := p.Match(lc)
		if err != nil {
			return err
		}
		if commitMatches {
			if keepGoing := onMatch(lc, highlights); !keepGoing {
				return nil
			}
		}
	}
	if cmdErr != nil {
		return cmdErr
	}
	return cs.Err()
}

type DiffFetcher struct {
	stdin   io.Writer
	stderr  bytes.Buffer
	scanner *bufio.Scanner
}

func StartDiffFetcher(ctx context.Context, dir string) (*DiffFetcher, error) {
	cmd := exec.CommandContext(ctx, "git", "diff-tree", "--stdin", "-p", "--format=format:")
	cmd.Dir = dir

	stdoutReader, stdoutWriter := io.Pipe()
	cmd.Stdout = stdoutWriter

	stdinReader, stdinWriter := io.Pipe()
	cmd.Stdin = stdinReader

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdoutReader)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if bytes.HasSuffix(data, []byte("\nENDOFPATCH\n")) {
			return len(data), data[:len(data)-len("ENDOFPATCH\n")], nil
		}

		return 0, nil, nil
	})

	return &DiffFetcher{
		stdin:   stdinWriter,
		scanner: scanner,
		stderr:  stderrBuf,
	}, nil
}

func (d *DiffFetcher) FetchDiff(hash []byte) ([]byte, error) {
	// HACK: There is no way (as far as I can tell) to make `git diff-tree --stdin` to
	// write a trailing null byte or tell us how much to read in advance, and since we're
	// using a long-running process, the stream doesn't close at the end, and we can't use the
	// start of a new patch to signify end of patch since we want to be able to do each round-trip
	// serially. We resort to sending the subprocess a bogus commit hash named "EOF", which it
	// will fail to read as a tree, and print back to stdout literally. We use this as a signal
	// that the subprocess is done outputting for this commit.
	d.stdin.Write(append(hash, []byte("\nENDOFPATCH\n")...))

	if d.scanner.Scan() {
		return d.scanner.Bytes(), nil
	} else if err := d.scanner.Err(); err != nil {
		return nil, err
	} else if d.stderr.String() != "" {
		return nil, errors.Errorf("git subprocess stderr: %s", d.stderr.String())
	}
	return nil, errors.New("expected scan to succeed")
}

type LazyCommit struct {
	*CommitView
	diffFetcher *DiffFetcher
}

func (l *LazyCommit) AuthorDate() (time.Time, error) {
	unixSeconds, err := strconv.Atoi(string(l.CommitView.AuthorDate))
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(unixSeconds), 0), nil
}

func (l *LazyCommit) Diff() (FormattedDiff, error) {
	gitDiff, err := l.diffFetcher.FetchDiff(l.Hash)
	if err != nil {
		return "", err
	}

	return FormatDiff(gitDiff), nil
}
