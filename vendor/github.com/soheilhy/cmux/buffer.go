package cmux

import "io"

var _ io.ReadWriter = (*buffer)(nil)

type buffer struct {
	read int
	data []byte
}

// From the io.Reader documentation:
//
// When Read encounters an error or end-of-file condition after
// successfully reading n > 0 bytes, it returns the number of
// bytes read.  It may return the (non-nil) error from the same call
// or return the error (and n == 0) from a subsequent call.
// An instance of this general case is that a Reader returning
// a non-zero number of bytes at the end of the input stream may
// return either err == EOF or err == nil.  The next Read should
// return 0, EOF.
//
// This function implements the latter behaviour, returning the
// (non-nil) error from the same call.
func (b *buffer) Read(p []byte) (int, error) {
	var err error
	n := copy(p, b.data[b.read:])
	b.read += n
	if b.read == len(b.data) {
		err = io.EOF
	}
	return n, err
}

func (b *buffer) Len() int {
	return len(b.data) - b.read
}

func (b *buffer) resetRead() {
	b.read = 0
}

// From the io.Writer documentation:
//
// Write writes len(p) bytes from p to the underlying data stream.
// It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early.
// Write must return a non-nil error if it returns n < len(p).
// Write must not modify the slice data, even temporarily.
//
// Implementations must not retain p.
//
// In a previous incarnation, this implementation retained the incoming slice.
func (b *buffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}
