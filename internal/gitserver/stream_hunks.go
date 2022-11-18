package gitserver

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const bufSize = 8

// blameHunkReader enables to read hunks from an io.Reader.
type blameHunkReader struct {
	rc io.ReadCloser

	hunksCh chan hunkResult
	buf     []*Hunk
	done    chan struct{}
}

func newBlameHunkReader(ctx context.Context, rc io.ReadCloser) HunkReader {
	br := &blameHunkReader{
		rc:      rc,
		hunksCh: make(chan hunkResult, bufSize),
		done:    make(chan struct{}),
		buf:     make([]*Hunk, 0, bufSize),
	}
	go br.readHunks(ctx)
	return br
}

type hunkResult struct {
	hunk *Hunk
	err  error
}

func (br *blameHunkReader) readHunks(ctx context.Context) {
	newHunkParser(br.rc, br.hunksCh, br.done).parse(ctx)
}

// Read returns a slice of hunks, along with a done boolean indicating if there is more to
// read.
func (br *blameHunkReader) Read() (hunks []*Hunk, done bool, err error) {
	for {
		select {
		case res := <-br.hunksCh:
			if res.err != nil {
				return nil, false, res.err
			}
			if len(br.buf) < bufSize {
				br.buf = append(br.buf, res.hunk)
				continue
			} else {
				hunks := br.buf
				br.buf = make([]*Hunk, 0, 8)
				return hunks, false, nil
			}
		case <-br.done:
			return nil, true, nil

		default:
			// if we're blocking on reads because git blame is slow, just send the first hunk we get ASAP.
			res := <-br.hunksCh
			return []*Hunk{res.hunk}, false, nil
		}
	}
}

type hunkParser struct {
	rc      io.ReadCloser
	sc      *bufio.Scanner
	hunksCh chan hunkResult
	done    chan struct{}

	// commits stores previously seen commits, so new hunks
	// whose annotations are abbreviated by git can still be
	// filled by the correct data even if the hunk entry doesn't
	// repeat them.
	commits map[string]*Hunk
}

func newHunkParser(rc io.ReadCloser, hunksCh chan hunkResult, done chan struct{}) hunkParser {
	return hunkParser{
		rc:      rc,
		hunksCh: hunksCh,
		done:    done,

		sc:      bufio.NewScanner(rc),
		commits: make(map[string]*Hunk),
	}
}

func (p hunkParser) parse(ctx context.Context) {
	defer p.rc.Close()

	var cur *Hunk
	for {
		if err := ctx.Err(); err != nil {
			close(p.done)
			break
		}

		// Do we have more to read?
		if !p.sc.Scan() {
			if cur != nil {
				if h, ok := p.commits[string(cur.CommitID)]; ok {
					cur.CommitID = h.CommitID
					cur.Author = h.Author
					cur.Message = h.Message
				}
				// If we have an ongoing entry, send it.
				p.hunksCh <- hunkResult{hunk: cur}
			}
			break
		}

		var done bool
		var err error
		// Read line from git blame, in porcelain format
		annotation, fields := p.scanLine()

		// On the first read, we have no hunk and the first thing we read is an entry.
		if cur == nil {
			cur, err = parseEntry(annotation, fields)
			if err != nil {
				p.hunksCh <- hunkResult{err: err}
			}
			continue
		}

		// After that, we're either reading extras, or a new entry.
		done, err = parseExtra(cur, annotation, fields)
		if err != nil {
			p.hunksCh <- hunkResult{err: err}
		}
		// If we've finished reading extras, we're looking at a new entry.
		if done {
			if h, ok := p.commits[string(cur.CommitID)]; ok {
				cur.CommitID = h.CommitID
				cur.Author = h.Author
				cur.Message = h.Message
			} else {
				p.commits[string(cur.CommitID)] = cur
			}

			p.hunksCh <- hunkResult{hunk: cur}

			cur, err = parseEntry(annotation, fields)
			if err != nil {
				p.hunksCh <- hunkResult{err: err}
			}
		}
	}

	// If there is an error from the scanner, send it back.
	if err := p.sc.Err(); err != nil {
		p.hunksCh <- hunkResult{err: err}
	}
	close(p.done)
}

// parseEntry turns a `67b7b725a7ff913da520b997d71c840230351e30 10 20 1` line from
// git blame into a hunk.
func parseEntry(rev string, content string) (*Hunk, error) {
	fields := strings.Split(content, " ")
	if len(fields) != 3 {
		return nil, errors.Errorf("HERE Expected at least 4 parts to hunkHeader, but got: '%s %s'", rev, content)
	}

	resultLine, _ := strconv.Atoi(fields[1])
	numLines, _ := strconv.Atoi(fields[2])

	return &Hunk{
		CommitID:  api.CommitID(rev),
		StartLine: resultLine,
		EndLine:   resultLine + numLines,
	}, nil
}

// parseExtra updates a hunk with data parsed from the other annotations such as `author ...`,
// `summary ...`.
func parseExtra(hunk *Hunk, annotation string, content string) (done bool, err error) {
	switch annotation {
	case "author":
		hunk.Author.Name = content
	case "author-mail":
		if len(content) >= 2 && content[0] == '<' && content[len(content)-1] == '>' {
			hunk.Author.Email = content[1 : len(content)-1]
		}
	case "author-time":
		var t int64
		t, err = strconv.ParseInt(content, 10, 64)
		hunk.Author.Date = time.Unix(t, 0).UTC()
	case "author-tz":
		// do nothing
	case "committer", "committer-mail", "committer-tz", "committer-time":
	case "summary":
		hunk.Message = content
	case "filename":
		hunk.Filename = content
	case "previous":
		// TODO
	case "boundary":
		// do nothing
	default:
		done = true
	}
	return
}

// scanLine reads a line from the scanner and returns the annotation along
// with the content, if any.
func (p hunkParser) scanLine() (string, string) {
	line := p.sc.Text()
	annotation, content, found := strings.Cut(line, " ")
	if found {
		return annotation, content
	}
	return line, ""
}

type mockHunkReader struct {
	hunks []*Hunk
	err   error

	idx int
}

func NewMockHunkReader(hunks []*Hunk, err error) HunkReader {
	return &mockHunkReader{
		hunks: hunks,
		err:   err,
	}
}

func (mh *mockHunkReader) Read() ([]*Hunk, bool, error) {
	if mh.err != nil {
		return nil, false, mh.err
	}
	if mh.idx < len(mh.hunks) {
		idx := mh.idx
		mh.idx++
		return []*Hunk{mh.hunks[idx]}, false, nil
	}
	return nil, true, nil
}
