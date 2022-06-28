package http

import (
	"bytes"
	"encoding/json"
)

// JSONArrayBuf builds up a JSON array by marshalling per item. Once the array
// has reached FlushSize it will be written out via Write and the buffer will
// be reset.
type JSONArrayBuf struct {
	FlushSize int
	Write     func([]byte) error

	buf bytes.Buffer
}

func NewJSONArrayBuf(flushSize int, write func([]byte) error) *JSONArrayBuf {
	b := &JSONArrayBuf{
		FlushSize: flushSize,
		Write:     write,
	}
	// Grow the buffer to flushSize to reduce the number of small allocations
	// caused by repeatedly growing the buffer
	b.buf.Grow(flushSize)
	return b
}

// Append marshals v and adds it to the json array buffer. If the size of the
// buffer exceed FlushSize the buffer is written out.
func (j *JSONArrayBuf) Append(v any) error {
	oldLen := j.buf.Len()

	if j.buf.Len() == 0 {
		j.buf.WriteByte('[')
	} else {
		j.buf.WriteByte(',')
	}

	enc := json.NewEncoder(&j.buf)
	if err := enc.Encode(v); err != nil {
		// Reset the buffer to where it was before failing to marshal
		j.buf.Truncate(oldLen)
		return err
	}

	// Trim the trailing newline left by the JSON encoder
	j.buf.Truncate(j.buf.Len() - 1)

	if j.buf.Len() >= j.FlushSize {
		return j.Flush()
	}
	return nil
}

// Flush writes and resets the buffer if there is data to write.
func (j *JSONArrayBuf) Flush() error {
	if j.buf.Len() == 0 {
		return nil
	}

	// Terminate array
	j.buf.WriteByte(']')

	buf := j.buf.Bytes()
	j.buf.Reset()
	return j.Write(buf)
}

func (j *JSONArrayBuf) Len() int {
	return j.buf.Len()
}
