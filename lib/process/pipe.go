package process

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os/exec"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PipeOutput reads stdout/stderr output of the given command into the two
// io.Writers.
//
// It returns a sync.WaitGroup. The caller *must* call the Wait() method of the
// WaitGroup after waiting for the *exec.Cmd to finish.
//
// See this issue for more details: https://github.com/golang/go/issues/21922
func PipeOutput(ctx context.Context, c *exec.Cmd, stdoutWriter, stderrWriter io.Writer) (*sync.WaitGroup, error) {
	return pipeOutputWithCopy(ctx, c, stdoutWriter, stderrWriter, func(w io.Writer, r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			fmt.Fprintln(w, scanner.Text())
		}
	})
}

func PipeOutputUnbuffered(ctx context.Context, c *exec.Cmd, stdoutWriter, stderrWriter io.Writer) (*sync.WaitGroup, error) {
	return pipeOutputWithCopy(ctx, c, stdoutWriter, stderrWriter, func(w io.Writer, r io.Reader) {
		_, err := io.Copy(w, r)
		// We can ignore ErrClosed because we get that if a process crashes
		if err != nil && !errors.Is(err, fs.ErrClosed) {
			panic(err)
		}
	})
}

func pipeOutputWithCopy(ctx context.Context, c *exec.Cmd, stdoutWriter, stderrWriter io.Writer, fn copyFunc) (*sync.WaitGroup, error) {
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

	wg := &sync.WaitGroup{}

	readIntoBuf := func(w io.Writer, r io.Reader) {
		defer wg.Done()

		fn(w, r)
	}

	wg.Add(2)
	go readIntoBuf(stdoutWriter, stdoutPipe)
	go readIntoBuf(stderrWriter, stderrPipe)

	return wg, nil
}

type copyFunc func(w io.Writer, r io.Reader)
