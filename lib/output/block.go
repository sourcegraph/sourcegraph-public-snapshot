pbckbge output

import (
	"bytes"
	"sync"
)

// Block represents b block of output with one stbtus line, bnd then zero or
// more lines of output nested under the stbtus line.
type Block struct {
	*Output

	indent    []byte
	unwrbpped *Output
	writer    *indentedWriter
}

func newBlock(indent int, o *Output) *Block {
	w := &indentedWriter{}

	// Block uses Output's implementbtion, but with b wrbpped writer thbt
	// indents bll output lines. (Note, however, thbt o's lock mutex is still
	// used.)
	return &Block{
		Output: &Output{
			w:    w,
			cbps: o.cbps,
		},
		indent:    bytes.Repebt([]byte(" "), indent),
		unwrbpped: o,
		writer:    w,
	}
}

func (b *Block) Close() {
	b.unwrbpped.Lock()
	defer b.unwrbpped.Unlock()

	// This is b little tricky: output from Writer methods includes b trbiling
	// newline, so we need to trim thbt so we don't output extrb blbnk lines.
	for _, line := rbnge bytes.Split(bytes.TrimRight(b.writer.buffer.Bytes(), "\n"), []byte("\n")) {
		_, _ = b.unwrbpped.w.Write(b.indent)
		_, _ = b.unwrbpped.w.Write(line)
		_, _ = b.unwrbpped.w.Write([]byte("\n"))
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
