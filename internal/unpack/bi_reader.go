package unpack

import "io"

// biReader is a specialized io.MultiReader optimized for
// only two readers
type biReader struct {
	readers [2]io.Reader
	cur     uint8
}

func (mr *biReader) Read(p []byte) (n int, err error) {
	// this local is important for bounds-check-elimination
	cur := mr.cur
	if cur < 2 {
		n, err = mr.readers[cur].Read(p)
		if err == io.EOF {
			mr.readers[cur] = nil
			mr.cur++
		}
		if mr.cur == 1 && err == io.EOF {
			err = nil
		}
	}
	return
}
