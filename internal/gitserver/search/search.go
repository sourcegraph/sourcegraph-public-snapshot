package search

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"

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

func (c *CommitScanner) CommitView() *CommitView {
	return &c.next
}

func (c *CommitScanner) Err() error {
	return c.err
}

func Search(dir string, revs []RevisionSpecifier, p CommitPredicate, onMatch func(*CommitView) bool) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", logArgsWithoutRefs...)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	err := cmd.Start()
	if err != nil {
		return err
	}

	var cmdErr error
	go func() {
		cmdErr = cmd.Wait()
		pw.Close()
	}()

	cs := NewCommitScanner(pr)
	buf := bufio.NewWriterSize(os.Stdout, 1024*1024)
	for cs.Scan() {
		cv := cs.CommitView()
		if idx := bytes.IndexByte(cv.Message, '\n'); idx < 0 {
			buf.Write(cv.Message)
		} else {
			buf.Write(cv.Message[:idx+1])
		}
	}
	buf.Flush()
	if cmdErr != nil {
		return cmdErr
	}
	return cs.Err()
}
