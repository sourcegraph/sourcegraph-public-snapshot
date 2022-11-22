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

// blameHunkReader enables to read hunks from an io.Reader.
type blameHunkReader struct {
	hunks chan hunkResult
}

func newBlameHunkReader(ctx context.Context, rc io.ReadCloser) HunkReader {
	br := &blameHunkReader{
		hunks: make(chan hunkResult),
	}
	go br.readHunks(ctx, rc)
	return br
}

type hunkResult struct {
	hunk *Hunk
	err  error
}

func (br *blameHunkReader) readHunks(ctx context.Context, rc io.ReadCloser) {
	newHunkParser(rc, br.hunks).parse(ctx)
}

// Read returns a slice of hunks, along with a done boolean indicating if there is more to
// read.
func (br *blameHunkReader) Read() ([]*Hunk, bool, error) {
	res, ok := <-br.hunks
	if !ok {
		return nil, true, nil
	}
	if res.err != nil {
		return nil, false, res.err
	} else {
		return []*Hunk{res.hunk}, false, nil
	}
}

type hunkParser struct {
	rc      io.ReadCloser
	sc      *bufio.Scanner
	hunksCh chan hunkResult

	// commits stores previously seen commits, so new hunks
	// whose annotations are abbreviated by git can still be
	// filled by the correct data even if the hunk entry doesn't
	// repeat them.
	commits map[string]*Hunk
}

func newHunkParser(rc io.ReadCloser, hunksCh chan hunkResult) hunkParser {
	return hunkParser{
		rc:      rc,
		hunksCh: hunksCh,

		sc:      bufio.NewScanner(rc),
		commits: make(map[string]*Hunk),
	}
}

// parse processes the output from git blame and sends hunks over p.hunksCh
// for p.Read() to consume. If an error is encountered, it will be sent to
// p.hunksCh and will stop reading.
//
// Because we do not control when p.Read is called, we have to account for
// the context being cancelled, to avoid leaking the goroutine running p.parse.
func (p hunkParser) parse(ctx context.Context) {
	defer p.rc.Close()
	defer close(p.hunksCh)

	var cur *Hunk
	for {
		if err := ctx.Err(); err != nil {
			return
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
				select {
				case p.hunksCh <- hunkResult{hunk: cur}:
				case <-ctx.Done():
					return
				}
			}
			break
		}

		// Read line from git blame, in porcelain format
		annotation, fields := p.scanLine()

		// On the first read, we have no hunk and the first thing we read is an entry.
		if cur == nil {
			var err error
			cur, err = parseEntry(annotation, fields)
			if err != nil {
				select {
				case p.hunksCh <- hunkResult{err: err}:
					return
				case <-ctx.Done():
					return
				}
			}
			continue
		}

		// After that, we're either reading extras, or a new entry.
		ok, err := parseExtra(cur, annotation, fields)
		if err != nil {
			select {
			case p.hunksCh <- hunkResult{err: err}:
				return
			case <-ctx.Done():
				return
			}
		}
		// If we've finished reading extras, we're looking at a new entry.
		if !ok {
			if h, ok := p.commits[string(cur.CommitID)]; ok {
				cur.CommitID = h.CommitID
				cur.Author = h.Author
				cur.Message = h.Message
			} else {
				p.commits[string(cur.CommitID)] = cur
			}

			select {
			case p.hunksCh <- hunkResult{hunk: cur}:
			case <-ctx.Done():
				return
			}

			cur, err = parseEntry(annotation, fields)
			if err != nil {
				select {
				case p.hunksCh <- hunkResult{err: err}:
					return
				case <-ctx.Done():
					return
				}
			}
		}
	}

	// If there is an error from the scanner, send it back.
	if err := p.sc.Err(); err != nil {
		select {
		case p.hunksCh <- hunkResult{err: err}:
			return
		case <-ctx.Done():
			return
		}
	}
}

// parseEntry turns a `67b7b725a7ff913da520b997d71c840230351e30 10 20 1` line from
// git blame into a hunk.
func parseEntry(rev string, content string) (*Hunk, error) {
	fields := strings.Split(content, " ")
	if len(fields) != 3 {
		return nil, errors.Errorf("Expected at least 4 parts to hunkHeader, but got: '%s %s'", rev, content)
	}

	resultLine, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, err
	}
	numLines, _ := strconv.Atoi(fields[2])
	if err != nil {
		return nil, err
	}

	return &Hunk{
		CommitID:  api.CommitID(rev),
		StartLine: resultLine,
		EndLine:   resultLine + numLines,
	}, nil
}

// parseExtra updates a hunk with data parsed from the other annotations such as `author ...`,
// `summary ...`.
func parseExtra(hunk *Hunk, annotation string, content string) (ok bool, err error) {
	ok = true
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
	case "boundary":
	default:
		// If it doesn't look like an entry, it's probably an unhandled git blame
		// annotation.
		if len(annotation) != 40 && len(strings.Split(content, " ")) != 3 {
			err = errors.Newf("unhandled git blame annotation: %s")
		}
		ok = false
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
