package search

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DiffFetcher is a handle to the stdin and stdout of a git diff-tree subprocess
// started with StartDiffFetcher
type DiffFetcher struct {
	dir string

	startOnce sync.Once
	stdin     io.Writer
	stderr    io.Reader
	scanner   *bufio.Scanner
	cancel    context.CancelFunc
	cmd       *exec.Cmd
}

// NewDiffFetcher starts a git diff-tree subprocess that waits, listening on stdin
// for comimt hashes to generate patches for.
func NewDiffFetcher(dir string) (*DiffFetcher, error) {

	return &DiffFetcher{dir: dir}, nil
}

func (d *DiffFetcher) Stop() {
	if d.cancel != nil {
		d.cancel()
		d.cmd.Wait()
	}
}

func (d *DiffFetcher) start() (err error) {
	d.startOnce.Do(func() {
		ctx := context.Background()
		ctx, d.cancel = context.WithCancel(ctx)
		d.cmd = exec.CommandContext(ctx, "git",
			"diff-tree",
			"--stdin",          // Read commit hashes from stdin
			"--no-prefix",      // Do not prefix file names with a/ and b/
			"-p",               // Output in patch format
			"--format=format:", // Output only the patch, not any other commit metadata
			"--root",           // Treat the root commit as a big creation event (otherwise the diff would be empty)
		)
		d.cmd.Dir = d.dir

		var stdoutReader io.ReadCloser
		stdoutReader, err = d.cmd.StdoutPipe()
		if err != nil {
			return
		}

		d.stdin, err = d.cmd.StdinPipe()
		if err != nil {
			return
		}

		d.stderr, err = d.cmd.StderrPipe()
		if err != nil {
			return
		}

		if err = d.cmd.Start(); err != nil {
			return
		}

		d.scanner = bufio.NewScanner(stdoutReader)
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
	})
	return err
}

// Fetch fetches a diff from the git diff-tree subprocess, writing to its stdin
// and waiting for its response on stdout. Note that this is not safe to call concurrently.
func (d *DiffFetcher) Fetch(hash []byte) ([]byte, error) {
	if err := d.start(); err != nil {
		return nil, err
	}
	// HACK: There is no way (as far as I can tell) to make `git diff-tree --stdin` to
	// write a trailing null byte or tell us how much to read in advance, and since we're
	// using a long-running process, the stream doesn't close at the end, and we can't use the
	// start of a new patch to signify end of patch since we want to be able to do each round-trip
	// serially. We resort to sending the subprocess a bogus commit hash named "ENDOFPATCH", which it
	// will fail to read as a tree, and print back to stdout literally. We use this as a signal
	// that the subprocess is done outputting for this commit.
	d.stdin.Write(append(hash, []byte("\nENDOFPATCH\n")...))

	if d.scanner.Scan() {
		return d.scanner.Bytes(), nil
	} else if err := d.scanner.Err(); err != nil {
		return nil, err
	} else if stderr, _ := io.ReadAll(d.stderr); len(stderr) > 0 {
		return nil, errors.Errorf("git subprocess stderr: %s", string(stderr))
	}
	return nil, errors.New("expected scan to succeed")
}
