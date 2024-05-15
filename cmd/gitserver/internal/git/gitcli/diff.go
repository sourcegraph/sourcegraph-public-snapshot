package gitcli

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) RawDiff(ctx context.Context, base string, head string, typ git.GitDiffComparisonType, paths ...string) (io.ReadCloser, error) {
	// baseOID, err := g.revParse(ctx, base)
	// if err != nil {
	// 	return nil, err
	// }
	// headOID, err := g.revParse(ctx, head)
	// if err != nil {
	// 	return nil, err
	// }

	return g.NewCommand(ctx, WithArguments(buildRawDiffArgs(api.CommitID(base), api.CommitID(head), typ, paths)...))
}

func buildRawDiffArgs(base, head api.CommitID, typ git.GitDiffComparisonType, paths []string) []string {
	var rangeType string
	switch typ {
	case git.GitDiffComparisonTypeIntersection:
		rangeType = "..."
	case git.GitDiffComparisonTypeOnlyInHead:
		rangeType = ".."
	}
	rangeSpec := string(base) + rangeType + string(head)

	return append([]string{
		"diff",
		"--find-renames",
		"--full-index",
		"--inter-hunk-context=3",
		"--no-prefix",
		rangeSpec,
		"--",
	}, paths...)
}

func (g *gitCLIBackend) ChangedFiles(ctx context.Context, base, head string) (git.ChangedFilesIterator, error) {
	args := []string{
		"diff-tree",
		"-r",
		"--root",
		"--format=format:",
		"--no-prefix",
		"--name-status",
		"--no-renames",
		"-z",
	}

	if base != "" {
		baseOID, err := g.revParse(ctx, base)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to resolve base commit %q", base)
		}

		args = append(args, string(baseOID))
	}

	headOID, err := g.revParse(ctx, head)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to resolve head commit %q", head)
	}

	args = append(args, string(headOID))

	rc, err := g.NewCommand(ctx, WithArguments(args...))
	if err != nil {
		return nil, errors.Wrap(err, "failed to run git diff-tree command")
	}

	return newGitDiffIterator(rc), nil
}

func newGitDiffIterator(rc io.ReadCloser) git.ChangedFilesIterator {
	scanner := bufio.NewScanner(rc)
	scanner.Split(byteutils.ScanNullLines)

	closeChan := make(chan struct{})
	closer := sync.OnceValue(func() error {
		err := rc.Close()
		close(closeChan)

		return err
	})

	return &gitDiffIterator{
		rc:             rc,
		scanner:        scanner,
		closeChan:      closeChan,
		onceFuncCloser: closer,
	}
}

type gitDiffIterator struct {
	rc      io.ReadCloser
	scanner *bufio.Scanner

	closeChan      chan struct{}
	onceFuncCloser func() error
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

func (i *gitDiffIterator) Close() error {
	return i.onceFuncCloser()
}

var _ git.ChangedFilesIterator = &gitDiffIterator{}

func (g *gitCLIBackend) DiffFetcher(ctx context.Context) (git.DiffFetcher, error) {
	d := &diffFetcher{}

	ctx, d.cancel = context.WithCancel(ctx)
	stdinR, stdinW := io.Pipe()
	// TODO: Close pipe.
	d.stdin = stdinW

	var stderr bytes.Buffer
	r, err := g.NewCommand(ctx,
		WithArguments(
			"diff-tree",
			"--stdin",          // Read commit hashes from stdin
			"--no-prefix",      // Do not prefix file names with a/ and b/
			"-p",               // Output in patch format
			"--format=format:", // Output only the patch, not any other commit metadata
			"--root",           // Treat the root commit as a big creation event (otherwise the diff would be empty)
		),
		WithStdin(stdinR),
		WithStderr(&stderr),
	)
	if err != nil {
		return nil, err
	}
	d.stderr = &stderr
	d.cmdReader = r

	// d.stderr, err = d.cmd.StderrPipe()
	// if err != nil {
	// 	return nil, err
	// }

	d.scanner = bufio.NewScanner(r)
	d.scanner.Buffer(make([]byte, 1024), 1<<30)
	d.scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Note that this only works when we write to stdin, then read from stdout before writing
		// anything else to stdin, since we are using `HasSuffix` and not `Contains`.
		if bytes.HasSuffix(data, []byte("ENDOFPATCH\n")) {
			if bytes.Equal(data, []byte("ENDOFPATCH\n")) {
				// Empty patch
				return len(data), data[:0], nil
			}
			return len(data), data[:len(data)-len("ENDOFPATCH\n")], nil
		}

		return 0, nil, nil
	})
	return d, nil
}

// diffFetcher is a handle to the stdin and stdout of a git diff-tree subprocess
// started with StartDiffFetcher
type diffFetcher struct {
	stdin     io.Writer
	stderr    *bytes.Buffer
	scanner   *bufio.Scanner
	cancel    context.CancelFunc
	cmdReader io.ReadCloser
}

func (d *diffFetcher) Close() error {
	if d.cancel != nil {
		d.cancel()
	}
	return d.cmdReader.Close()
}

func (d *diffFetcher) Fetch(sha api.CommitID) (io.Reader, error) {
	// TODO: Maybe return a revisionnotfounderror here when the SHA doesn't exist.

	// Check if there was any error. TODO: Maybe do that before returning io.EOF
	// from the returned reader instead?
	if stderr := d.stderr.Bytes(); len(stderr) > 0 {
		return nil, errors.Errorf("git subprocess stderr: %s", string(stderr))
	}

	// HACK: There is no way (as far as I can tell) to make `git diff-tree --stdin` to
	// write a trailing null byte or tell us how much to read in advance, and since we're
	// using a long-running process, the stream doesn't close at the end, and we can't use the
	// start of a new patch to signify end of patch since we want to be able to do each round-trip
	// serially. We resort to sending the subprocess a bogus commit hash named "ENDOFPATCH", which it
	// will fail to read as a tree, and print back to stdout literally. We use this as a signal
	// that the subprocess is done outputting for this commit.
	if _, err := d.stdin.Write(append([]byte(sha), []byte("\nENDOFPATCH\n")...)); err != nil {
		return nil, err
	}

	// TODO: Don't use this scanner, it reads an entire diff into memory.
	if d.scanner.Scan() {
		return bytes.NewReader(d.scanner.Bytes()), nil
	} else if err := d.scanner.Err(); err != nil {
		return nil, err
	} else if stderr, _ := io.ReadAll(d.stderr); len(stderr) > 0 {
		return nil, errors.Errorf("git subprocess stderr: %s", string(stderr))
	}
	return nil, errors.New("expected scan to succeed")
	// return splitReader(d.cmdReader, []byte("\nENDOFPATCH\n")), nil
}

// splitReader takes in an io.Reader and a separator, and returns an io.Reader that will
// forward reads to r until it encounters the separator, at which point it will stop reading
// from r and returns io.EOF.
func splitReader(r io.Reader, sep []byte) io.Reader {
	return &splittedReader{r: r, sep: sep, buf: make([]byte, 0, len(sep))}
}

type splittedReader struct {
	r   io.Reader
	sep []byte
	buf []byte
}

func (sr *splittedReader) Read(p []byte) (int, error) {
	n := 0
	for len(p) > 0 {
		if len(sr.buf) > 0 {
			m := copy(p, sr.buf)
			p = p[m:]
			n += m
			sr.buf = sr.buf[m:]
			if len(sr.buf) == 0 {
				return n, io.EOF
			}
			continue
		}

		buf := make([]byte, len(p))
		m, err := sr.r.Read(buf)
		if err != nil && err != io.EOF {
			return n, err
		}
		buf = buf[:m]

		idx := bytes.Index(buf, sr.sep)
		if idx != -1 {
			m = copy(p, buf[:idx])
			sr.buf = append(sr.buf, buf[idx+len(sr.sep):]...)
			n += m
			return n, io.EOF
		}

		m = copy(p, buf)
		p = p[m:]
		n += m

		if err == io.EOF {
			return n, io.EOF
		}
	}

	return n, nil
}
