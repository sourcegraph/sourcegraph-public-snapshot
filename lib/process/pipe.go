package process

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"

	"golang.org/x/sync/errgroup"

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

// PipeOutput reads stdout/stderr output of the given command into the two
// io.Writers.
//
// It returns a errgroup.Group. The caller *must* call the Wait() method of the
// errgroup.Group after waiting for the *exec.Cmd to finish.
//
// See this issue for more details: https://github.com/golang/go/issues/21922
func PipeOutput(ctx context.Context, c cmdPiper, stdoutWriter, stderrWriter io.Writer) (*errgroup.Group, error) {
	pipe := func(w io.Writer, r io.Reader) error {
		scanner := bufio.NewScanner(r)
		scanner.Split(scanLinesWithNewline)

		buf := make([]byte, initialBufSize)
		scanner.Buffer(buf, maxTokenSize)

		for scanner.Scan() {
			fmt.Fprint(w, scanner.Text())
		}

		return scanner.Err()
	}

	return pipeProcessOutput(ctx, c, stdoutWriter, stderrWriter, pipe)
}

// PipeOutputUnbuffered is the unbuffered version of PipeOutput and uses
// io.Copy instead of piping output line-based to the output.
func PipeOutputUnbuffered(ctx context.Context, c cmdPiper, stdoutWriter, stderrWriter io.Writer) (*errgroup.Group, error) {
	pipe := func(w io.Writer, r io.Reader) error {
		_, err := io.Copy(w, r)
		// We can ignore ErrClosed because we get that if a process crashes
		if err != nil && !errors.Is(err, fs.ErrClosed) {
			return err
		}
		return nil
	}

	return pipeProcessOutput(ctx, c, stdoutWriter, stderrWriter, pipe)
}

func pipeProcessOutput(ctx context.Context, c cmdPiper, stdoutWriter, stderrWriter io.Writer, fn pipe) (*errgroup.Group, error) {
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderrPipe, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		// We start a goroutine here to make sure that our pipes are closed
		// when the context is canceled.
		//
		// See enterprise/cmd/executor/internal/command/run.go for more details.
		<-ctx.Done()
		stdoutPipe.Close()
		stderrPipe.Close()
	}()

	eg := &errgroup.Group{}

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
