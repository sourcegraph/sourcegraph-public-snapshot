package output

import "bytes"

// Block represents a block of output with one status line, and then zero or
// more lines of output nested under the status line.
type Block struct {
	*Output
}

func newBlock(indent int, o *Output) *Block {
	// Block uses Output's implementation, but with a wrapped writer that
	// indents all output lines. (Note, however, that o's lock mutex is still
	// used.)
	return &Block{
		&Output{
			w: &indentedWriter{
				o:      o,
				indent: bytes.Repeat([]byte(" "), indent),
			},
			caps: o.caps,
			opts: o.opts,
		},
	}
}

type indentedWriter struct {
	o      *Output
	indent []byte
}

func (w *indentedWriter) Write(p []byte) (int, error) {
	w.o.lock.Lock()
	defer w.o.lock.Unlock()

	// This is a little tricky: output from Writer methods includes a trailing
	// newline, so we need to trim that so we don't output extra blank lines.
	for _, line := range bytes.Split(bytes.TrimRight(p, "\n"), []byte("\n")) {
		w.o.w.Write(w.indent)
		w.o.w.Write(line)
		w.o.w.Write([]byte("\n"))
	}

	return len(p), nil
}
