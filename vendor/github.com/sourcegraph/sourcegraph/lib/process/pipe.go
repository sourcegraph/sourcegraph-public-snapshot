package process

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// initialBufSize is the initial size of the buffer that PipeOutput uses to
// read lines.
const initialBufSize = 4 * 1024 // 4k
// maxTokenSize is the max size of a token that PipeOutput reads.
const maxTokenSize = 100 * 1024 * 1024 // 100mb

type pipe func(w io.Writer, r io.Reader) error

type cmdPiper interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}

// NewOutputScannerWithSplit creates a new bufio.Scanner using the given split
// function with well-working defaults for the initial and max buf sizes.
func NewOutputScannerWithSplit(r io.Reader, split bufio.SplitFunc) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(split)
	buf := make([]byte, initialBufSize)
	scanner.Buffer(buf, maxTokenSize)
	return scanner
}

// PipeOutput reads stdout/stderr output of the given command into the two
// io.Writers.
//
// It returns a errgroup.Group. The caller *must* call the Wait() method of the
// errgroup.Group **before** waiting for the *exec.Cmd to finish.
//
// The passed in context should be canceled when done.
//
// See this issue for more details: https://github.com/golang/go/issues/21922
func PipeOutput(ctx context.Context, c cmdPiper, stdoutWriter, stderrWriter io.Writer) (*pool.ErrorPool, error) {
	pipe := func(w io.Writer, r io.Reader) error {
		scanner := NewOutputScannerWithSplit(r, scanLinesWithNewline)

		for scanner.Scan() {
			if _, err := fmt.Fprint(w, scanner.Text()); err != nil {
				return err
			}
		}

		return scanner.Err()
	}

	return PipeProcessOutput(ctx, c, stdoutWriter, stderrWriter, pipe)
}

// PipeOutputUnbuffered is the unbuffered version of PipeOutput and uses
// io.Copy instead of piping output line-based to the output.
func PipeOutputUnbuffered(ctx context.Context, c cmdPiper, stdoutWriter, stderrWriter io.Writer) (*pool.ErrorPool, error) {
	pipe := func(w io.Writer, r io.Reader) error {
		_, err := io.Copy(w, r)
		// We can ignore ErrClosed because we get that if a process crashes
		if err != nil && !errors.Is(err, fs.ErrClosed) {
			return err
		}
		return nil
	}

	return PipeProcessOutput(ctx, c, stdoutWriter, stderrWriter, pipe)
}

func PipeProcessOutput(ctx context.Context, c cmdPiper, stdoutWriter, stderrWriter io.Writer, fn pipe) (*pool.ErrorPool, error) {
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to attach stdout pipe")
	}

	stderrPipe, err := c.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to attach stderr pipe")
	}

	context.AfterFunc(ctx, func() {
		// There is a deadlock condition due the following strange decisions:
		//
		// 1. The pipes attached to a command are not closed if the context
		//    attached to the command is canceled. The pipes are only closed
		//    after Wait has been called.
		// 2. According to the docs, we are not meant to call cmd.Wait() until
		//    we have complete read the pipes attached to the command.
		//
		// Since we're following the expected usage, we block on a wait group
		// tracking the consumption of stdout and stderr pipes in two separate
		// goroutines between calls to Start and Wait. This means that if there
		// is a reason the command is abandoned but the pipes are not closed
		// (such as context cancellation), we will hang indefinitely.
		//
		// To be defensive, we'll forcibly close both pipes when the context has
		// finished. These may return an ErrClosed condition, but we don't really
		// care: the command package doesn't surface errors when closing the pipes
		// either.
		stdoutPipe.Close()
		stderrPipe.Close()
	})

	eg := pool.New().WithErrors()

	eg.Go(func() error { return fn(stdoutWriter, stdoutPipe) })
	eg.Go(func() error { return fn(stderrWriter, stderrPipe) })

	return eg, nil
}

// scanLinesWithNewline is a modified version of bufio.ScanLines that retains
// the trailing newline byte(s) in the returned token.
func scanLinesWithNewline(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
