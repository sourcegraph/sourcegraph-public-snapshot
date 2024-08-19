package output

import (
	"bytes"
	"sync"
)

// Block represents a block of output with one status line, and then zero or
// more lines of output nested under the status line.
type Block struct {
	*Output

	indent    []byte
	unwrapped *Output
	writer    *indentedWriter
}

func newBlock(indent int, o *Output) *Block {
	w := &indentedWriter{}

	// Block uses Output's implementation, but with a wrapped writer that
	// indents all output lines. (Note, however, that o's lock mutex is still
	// used.)
	return &Block{
		Output: &Output{
			w:    w,
			caps: o.caps,
		},
		indent:    bytes.Repeat([]byte(" "), indent),
		unwrapped: o,
		writer:    w,
	}
}

func (b *Block) Close() {
	b.unwrapped.Lock()
	defer b.unwrapped.Unlock()

	// This is a little tricky: output from Writer methods includes a trailing
	// newline, so we need to trim that so we don't output extra blank lines.
	for _, line := range bytes.Split(bytes.TrimRight(b.writer.buffer.Bytes(), "\n"), []byte("\n")) {
		_, _ = b.unwrapped.w.Write(b.indent)
		_, _ = b.unwrapped.w.Write(line)
		_, _ = b.unwrapped.w.Write([]byte("\n"))
	}
}

type indentedWriter struct {
	buffer bytes.Buffer
	lock   sync.Mutex
}

func (w *indentedWriter) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	return w.buffer.Write(p)
}
