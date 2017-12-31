package gophermail

import (
	"bytes"
	"encoding/base64"
	"io"
)

var delimiter = []byte("\r\n") // CRLF

// Lines should be no longer than 76 characters excluding the CRLF.
// See RFC 2045 and GitHub Issue #13.
const maxLength = 76

// splittingWriter is an io.WriteCloser that delimits
// the written data into fixed length chunks using
// a separator character sequence.
type splittingWriter struct {
	b       *bytes.Buffer
	w       io.Writer
	flushed bool
}

func (t *splittingWriter) Write(p []byte) (n int, err error) {

	// Write everything to our buffer.
	n, err = t.b.Write(p)

	// Check if our buffer can be flushed.
	for t.b.Len() >= maxLength {
		// If this isn't the first time flushing
		// the buffer then we need to write a delimiter.
		if t.flushed {
			_, err = t.w.Write(delimiter)
			if err != nil {
				return
			}
		}

		// Copy the next chunk out of the buffer.
		var n2 int64
		n2, err = io.CopyN(t.w, t.b, maxLength)
		if n2 > 0 {
			t.flushed = true
		}
	}

	return
}

func (t *splittingWriter) Close() (err error) {
	// Flush everything in the buffer.
	if t.b.Len() > 0 {
		// If this isn't the first time flushing
		// the buffer then we need to write a delimiter.
		if t.flushed {
			_, err = t.w.Write(delimiter)
			if err != nil {
				return
			}
		}

		var n int64
		n, err = io.Copy(t.w, t.b)
		if n > 0 {
			t.flushed = true
		}
	}

	return
}

type base64MimeEncoder struct {
	enc io.WriteCloser
	w   io.WriteCloser
}

func (t *base64MimeEncoder) Write(p []byte) (n int, err error) {
	n, err = t.enc.Write(p)
	return
}

func (t *base64MimeEncoder) Close() (err error) {
	err = t.enc.Close()
	if err != nil {
		return err
	}

	err = t.w.Close()
	return
}

func NewBase64MimeEncoder(w io.Writer) io.WriteCloser {
	splitter := &splittingWriter{
		w: w,
		b: &bytes.Buffer{},
	}
	t := &base64MimeEncoder{w: splitter}
	t.enc = base64.NewEncoder(base64.StdEncoding, splitter)
	return t
}
