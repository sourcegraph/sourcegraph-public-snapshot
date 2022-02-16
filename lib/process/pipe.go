package process

import (
	"context"
	"io"
	"io/fs"
	"os/exec"

	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PipeOutput reads stdout/stderr output of the given command into the two
// io.Writers.
//
// It returns a sync.WaitGroup. The caller *must* call the Wait() method of the
// WaitGroup after waiting for the *exec.Cmd to finish.
//
// See this issue for more details: https://github.com/golang/go/issues/21922
func PipeOutputUnbuffered(ctx context.Context, c *exec.Cmd, stdoutWriter, stderrWriter io.Writer) (*errgroup.Group, error) {
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

	readIntoBuf := func(w io.Writer, r io.Reader) error {
		_, err := io.Copy(w, r)
		// We can ignore ErrClosed because we get that if a process crashes
		if err != nil && !errors.Is(err, fs.ErrClosed) {
			return err
		}
		return nil
	}

	eg.Go(func() error {
		return readIntoBuf(stdoutWriter, stdoutPipe)
	})
	eg.Go(func() error {
		return readIntoBuf(stderrWriter, stderrPipe)
	})
	return eg, nil
}
