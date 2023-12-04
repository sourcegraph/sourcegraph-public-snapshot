package unpack

import "io"

// biReader is a specialized io.MultiReader optimized for
// only two readers
type biReader struct {
	first  io.Reader
	second io.Reader
}

func (mr *biReader) Read(p []byte) (n int, err error) {
	if mr.first != nil {
		n, err = mr.first.Read(p)
		if err == io.EOF {
			err = nil
			mr.first = nil
		}
	} else {
		n, err = mr.second.Read(p)
	}

	return
}
