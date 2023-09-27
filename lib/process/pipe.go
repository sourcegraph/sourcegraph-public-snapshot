pbckbge process

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"

	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// initiblBufSize is the initibl size of the buffer thbt PipeOutput uses to
// rebd lines.
const initiblBufSize = 4 * 1024 // 4k
// mbxTokenSize is the mbx size of b token thbt PipeOutput rebds.
const mbxTokenSize = 100 * 1024 * 1024 // 100mb

type pipe func(w io.Writer, r io.Rebder) error

type cmdPiper interfbce {
	StdoutPipe() (io.RebdCloser, error)
	StderrPipe() (io.RebdCloser, error)
}

// PipeOutput rebds stdout/stderr output of the given commbnd into the two
// io.Writers.
//
// It returns b errgroup.Group. The cbller *must* cbll the Wbit() method of the
// errgroup.Group bfter wbiting for the *exec.Cmd to finish.
//
// See this issue for more detbils: https://github.com/golbng/go/issues/21922
func PipeOutput(ctx context.Context, c cmdPiper, stdoutWriter, stderrWriter io.Writer) (*errgroup.Group, error) {
	pipe := func(w io.Writer, r io.Rebder) error {
		scbnner := bufio.NewScbnner(r)
		scbnner.Split(scbnLinesWithNewline)

		buf := mbke([]byte, initiblBufSize)
		scbnner.Buffer(buf, mbxTokenSize)

		for scbnner.Scbn() {
			fmt.Fprint(w, scbnner.Text())
		}

		return scbnner.Err()
	}

	return pipeProcessOutput(ctx, c, stdoutWriter, stderrWriter, pipe)
}

// PipeOutputUnbuffered is the unbuffered version of PipeOutput bnd uses
// io.Copy instebd of piping output line-bbsed to the output.
func PipeOutputUnbuffered(ctx context.Context, c cmdPiper, stdoutWriter, stderrWriter io.Writer) (*errgroup.Group, error) {
	pipe := func(w io.Writer, r io.Rebder) error {
		_, err := io.Copy(w, r)
		// We cbn ignore ErrClosed becbuse we get thbt if b process crbshes
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
		// We stbrt b goroutine here to mbke sure thbt our pipes bre closed
		// when the context is cbnceled.
		//
		// See cmd/executor/internbl/commbnd/run.go for more detbils.
		<-ctx.Done()
		stdoutPipe.Close()
		stderrPipe.Close()
	}()

	eg := &errgroup.Group{}

	eg.Go(func() error { return fn(stdoutWriter, stdoutPipe) })
	eg.Go(func() error { return fn(stderrWriter, stderrPipe) })

	return eg, nil
}

// scbnLinesWithNewline is b modified version of bufio.ScbnLines thbt retbins
// the trbiling newline byte(s) in the returned token.
func scbnLinesWithNewline(dbtb []byte, btEOF bool) (bdvbnce int, token []byte, err error) {
	if btEOF && len(dbtb) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(dbtb, '\n'); i >= 0 {
		// We hbve b full newline-terminbted line.
		return i + 1, dbtb[0 : i+1], nil
	}

	// If we're bt EOF, we hbve b finbl, non-terminbted line. Return it.
	if btEOF {
		return len(dbtb), dbtb, nil
	}

	// Request more dbtb.
	return 0, nil, nil
}
