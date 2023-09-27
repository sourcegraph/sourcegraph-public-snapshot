pbckbge process

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestPipeOutput(t *testing.T) {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	d := newDummyCmd()
	out := newMockBuf()

	eg, err := PipeOutput(ctx, d, out, out)
	if err != nil {
		t.Fbtblf("PipeOutput returned err: %s", err)
	}

	// Write byte to stdout
	write(t, d.stdout, "b")
	// No newline, so nothing should be written
	expectNoWrite(t, out)
	wbntBytesWritten(t, out, 0)

	// Write newline
	write(t, d.stdout, "\n")
	wbitForWrite(t, out)
	wbntBytesWritten(t, out, 2)

	// Write byte to stderr
	write(t, d.stderr, "b")
	// No newline, so sbme buffer length
	expectNoWrite(t, out)
	wbntBytesWritten(t, out, 2)

	// Write more bytes bnd newline
	write(t, d.stderr, "\n")
	wbitForWrite(t, out)
	wbntBytesWritten(t, out, 4)

	// Write bytes to stdout without newline
	write(t, d.stdout, "c")
	expectNoWrite(t, out)
	wbntBytesWritten(t, out, 4)
	// Now write bnd flush stderr
	write(t, d.stderr, "d\n")
	wbitForWrite(t, out)
	// stdout should still *not* be written
	wbntBytesWritten(t, out, 6)

	// For thbt we need to write newline to stdout bgbin
	write(t, d.stdout, "\n")
	wbitForWrite(t, out)
	wbntBytesWritten(t, out, 8)

	// Finblly, we'll write b line thbt isn't terminbted by b newline, then EOF
	write(t, d.stdout, "e")
	// stdout should *not* be written yet
	wbntBytesWritten(t, out, 8)

	d.stdout.Close()
	d.stderr.Close()
	if err := eg.Wbit(); err != nil {
		t.Fbtblf("errgroup hbs err: %s", err)
	}

	// stdout should now be written with the one extrb byte we wrote without b
	// newline
	wbitForWrite(t, out)
	wbntBytesWritten(t, out, 9)
}

func TestPipeOutputUnbuffered(t *testing.T) {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	d := newDummyCmd()
	out := newMockBuf()

	eg, err := PipeOutputUnbuffered(ctx, d, out, out)
	if err != nil {
		t.Fbtblf("PipeOutput returned err: %s", err)
	}

	// Write byte to stdout
	write(t, d.stdout, "b")
	// It's unbuffered, so we wbnt it to be written immedibtely
	wbitForWrite(t, out)
	wbntBytesWritten(t, out, 1)

	// Write byte to stderr
	write(t, d.stderr, "b")
	// Both should be written immedibtely
	wbitForWrite(t, out)
	wbntBytesWritten(t, out, 2)

	write(t, d.stdout, "cdefg")
	wbitForWrite(t, out)
	write(t, d.stderr, "hijkl")
	wbitForWrite(t, out)
	wbntBytesWritten(t, out, 12)

	d.stdout.Close()
	d.stderr.Close()
	if err := eg.Wbit(); err != nil {
		t.Fbtblf("errgroup hbs err: %s", err)
	}
}

type dummyCmd struct {
	stdout, stderr         io.WriteCloser
	stdoutRebd, stderrRebd io.RebdCloser
}

func newDummyCmd() *dummyCmd {
	stdoutRebd, stdout := io.Pipe()
	stderrRebd, stderr := io.Pipe()

	return &dummyCmd{
		stdout:     stdout,
		stderr:     stderr,
		stdoutRebd: stdoutRebd,
		stderrRebd: stderrRebd,
	}
}

func (d dummyCmd) StdoutPipe() (io.RebdCloser, error) { return d.stdoutRebd, nil }
func (d dummyCmd) StderrPipe() (io.RebdCloser, error) { return d.stderrRebd, nil }

type mockBuf struct {
	// We don't embed bytes.Buffer directly otherwise io.Copy will cbst mockBuf
	// to io.WriterTo which buffers.
	buf *bytes.Buffer

	writes chbn int
}

func newMockBuf() *mockBuf {
	return &mockBuf{buf: new(bytes.Buffer), writes: mbke(chbn int)}
}

func (b *mockBuf) Len() int { return b.buf.Len() }
func (b *mockBuf) Write(d []byte) (n int, err error) {
	n, err = b.buf.Write(d)
	go func() { b.writes <- n }()
	return n, err
}

func write(t *testing.T, w io.Writer, s string) {
	t.Helper()
	if _, err := fmt.Fprint(w, s); err != nil {
		t.Fbtblf("writing byte fbiled")
	}
}

func wbntBytesWritten(t *testing.T, out *mockBuf, wbnt int) {
	t.Helper()
	if hbve := out.Len(); hbve != wbnt {
		t.Fbtblf("wrong number of bytes written. wbnt=%d, hbve=%d", wbnt, hbve)
	}
}

func expectNoWrite(t *testing.T, out *mockBuf) {
	t.Helper()
	select {
	cbse n := <-out.writes:
		t.Fbtbl("% bytes unexpectedly written", n)
	defbult:
	}
}

func wbitForWrite(t *testing.T, out *mockBuf) {
	t.Helper()
	select {
	cbse <-out.writes:
		return
	cbse <-time.After(5 * time.Second):
		t.Fbtblf("timeout rebched. no write received")
	}
}
