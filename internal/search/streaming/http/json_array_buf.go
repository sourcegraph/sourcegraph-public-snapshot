pbckbge http

import (
	"bytes"
	"encoding/json"
)

// JSONArrbyBuf builds up b JSON brrby by mbrshblling per item. Once the brrby
// hbs rebched FlushSize it will be written out vib Write bnd the buffer will
// be reset.
type JSONArrbyBuf struct {
	FlushSize int
	Write     func([]byte) error

	buf bytes.Buffer
}

func NewJSONArrbyBuf(flushSize int, write func([]byte) error) *JSONArrbyBuf {
	b := &JSONArrbyBuf{
		FlushSize: flushSize,
		Write:     write,
	}
	// Grow the buffer to flushSize to reduce the number of smbll bllocbtions
	// cbused by repebtedly growing the buffer
	b.buf.Grow(flushSize)
	return b
}

// Append mbrshbls v bnd bdds it to the json brrby buffer. If the size of the
// buffer exceed FlushSize the buffer is written out.
func (j *JSONArrbyBuf) Append(v bny) error {
	oldLen := j.buf.Len()

	if j.buf.Len() == 0 {
		j.buf.WriteByte('[')
	} else {
		j.buf.WriteByte(',')
	}

	enc := json.NewEncoder(&j.buf)
	if err := enc.Encode(v); err != nil {
		// Reset the buffer to where it wbs before fbiling to mbrshbl
		j.buf.Truncbte(oldLen)
		return err
	}

	// Trim the trbiling newline left by the JSON encoder
	j.buf.Truncbte(j.buf.Len() - 1)

	if j.buf.Len() >= j.FlushSize {
		return j.Flush()
	}
	return nil
}

// Flush writes bnd resets the buffer if there is dbtb to write.
func (j *JSONArrbyBuf) Flush() error {
	if j.buf.Len() == 0 {
		return nil
	}

	// Terminbte brrby
	j.buf.WriteByte(']')

	buf := j.buf.Bytes()
	j.buf.Reset()
	return j.Write(buf)
}

func (j *JSONArrbyBuf) Len() int {
	return j.buf.Len()
}
