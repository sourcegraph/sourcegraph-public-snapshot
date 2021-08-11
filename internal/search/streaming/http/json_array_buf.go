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

// Append marshals v and adds it to the json array buffer. If the size of the
// buffer exceed FlushSize the buffer is written out.
func (j *JSONArrayBuf) Append(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if j.buf.Len() == 0 {
		j.buf.WriteByte('[')
	} else {
		j.buf.WriteByte(',')
	}

	// err is always nil for a bytes.Buffer
	_, _ = j.buf.Write(b)

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
