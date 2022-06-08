package process

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestPipeOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := newDummyCmd()
	out := newMockBuf()

	eg, err := PipeOutput(ctx, d, out, out)
	if err != nil {
		t.Fatalf("PipeOutput returned err: %s", err)
	}

	// Write byte to stdout
	write(t, d.stdout, "a")
	// No newline, so nothing should be written
	expectNoWrite(t, out)
	wantBytesWritten(t, out, 0)

	// Write newline
	write(t, d.stdout, "\n")
	waitForWrite(t, out)
	wantBytesWritten(t, out, 2)

	// Write byte to stderr
	write(t, d.stderr, "b")
	// No newline, so same buffer length
	expectNoWrite(t, out)
	wantBytesWritten(t, out, 2)

	// Write more bytes and newline
	write(t, d.stderr, "\n")
	waitForWrite(t, out)
	wantBytesWritten(t, out, 4)

	// Write bytes to stdout without newline
	write(t, d.stdout, "c")
	expectNoWrite(t, out)
	wantBytesWritten(t, out, 4)
	// Now write and flush stderr
	write(t, d.stderr, "d\n")
	waitForWrite(t, out)
	// stdout should still *not* be written
	wantBytesWritten(t, out, 6)

	// For that we need to write newline to stdout again
	write(t, d.stdout, "\n")
	waitForWrite(t, out)
	wantBytesWritten(t, out, 8)

	// Finally, we'll write a line that isn't terminated by a newline, then EOF
	write(t, d.stdout, "e")
	// stdout should *not* be written yet
	wantBytesWritten(t, out, 8)

	d.stdout.Close()
	d.stderr.Close()
	if err := eg.Wait(); err != nil {
		t.Fatalf("errgroup has err: %s", err)
	}

	// stdout should now be written with the one extra byte we wrote without a
	// newline
	waitForWrite(t, out)
	wantBytesWritten(t, out, 9)
}

func TestPipeOutputUnbuffered(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := newDummyCmd()
	out := newMockBuf()

	eg, err := PipeOutputUnbuffered(ctx, d, out, out)
	if err != nil {
		t.Fatalf("PipeOutput returned err: %s", err)
	}

	// Write byte to stdout
	write(t, d.stdout, "a")
	// It's unbuffered, so we want it to be written immediately
	waitForWrite(t, out)
	wantBytesWritten(t, out, 1)

	// Write byte to stderr
	write(t, d.stderr, "b")
	// Both should be written immediately
	waitForWrite(t, out)
	wantBytesWritten(t, out, 2)

	write(t, d.stdout, "cdefg")
	waitForWrite(t, out)
	write(t, d.stderr, "hijkl")
	waitForWrite(t, out)
	wantBytesWritten(t, out, 12)

	d.stdout.Close()
	d.stderr.Close()
	if err := eg.Wait(); err != nil {
		t.Fatalf("errgroup has err: %s", err)
	}
}

type dummyCmd struct {
	stdout, stderr         io.WriteCloser
	stdoutRead, stderrRead io.ReadCloser
}

func newDummyCmd() *dummyCmd {
	stdoutRead, stdout := io.Pipe()
	stderrRead, stderr := io.Pipe()

	return &dummyCmd{
		stdout:     stdout,
		stderr:     stderr,
		stdoutRead: stdoutRead,
		stderrRead: stderrRead,
	}
}

func (d dummyCmd) StdoutPipe() (io.ReadCloser, error) { return d.stdoutRead, nil }
func (d dummyCmd) StderrPipe() (io.ReadCloser, error) { return d.stderrRead, nil }

type mockBuf struct {
	// We don't embed bytes.Buffer directly otherwise io.Copy will cast mockBuf
	// to io.WriterTo which buffers.
	buf *bytes.Buffer

	writes chan int
}

func newMockBuf() *mockBuf {
	return &mockBuf{buf: new(bytes.Buffer), writes: make(chan int)}
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
		t.Fatalf("writing byte failed")
	}
}

func wantBytesWritten(t *testing.T, out *mockBuf, want int) {
	t.Helper()
	if have := out.Len(); have != want {
		t.Fatalf("wrong number of bytes written. want=%d, have=%d", want, have)
	}
}

func expectNoWrite(t *testing.T, out *mockBuf) {
	t.Helper()
	select {
	case n := <-out.writes:
		t.Fatal("% bytes unexpectedly written", n)
	default:
	}
}

func waitForWrite(t *testing.T, out *mockBuf) {
	t.Helper()
	select {
	case <-out.writes:
		return
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout reached. no write received")
	}
}
