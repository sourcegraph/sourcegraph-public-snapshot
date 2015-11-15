package cmux

import "io"

type buffer struct {
	read int
	data []byte
}

func (b *buffer) Read(p []byte) (n int, err error) {
	n = len(b.data) - b.read
	if n == 0 {
		return 0, io.EOF
	}

	if len(p) < n {
		n = len(p)
	}

	copy(p[:n], b.data[b.read:b.read+n])
	b.read += n
	return
}

func (b *buffer) Len() int {
	return len(b.data) - b.read
}

func (b *buffer) resetRead() {
	b.read = 0
}

func (b *buffer) Write(p []byte) (n int, err error) {
	n = len(p)
	if b.data == nil {
		b.data = p[:n:n]
		return
	}

	b.data = append(b.data, p...)
	return
}
