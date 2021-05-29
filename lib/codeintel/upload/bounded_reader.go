package upload

import "io"

type boundedReader struct {
	readerAt  io.ReaderAt
	minOffset int64
	maxOffset int64
}

// newBoundedReader bounds the given reader between the given min and max byte offsets.
func newBoundedReader(r io.ReaderAt, minOffset, maxOffset int64) io.Reader {
	return &boundedReader{
		readerAt:  r,
		minOffset: minOffset,
		maxOffset: maxOffset,
	}
}

func (r *boundedReader) Read(buf []byte) (int, error) {
	if r.minOffset >= r.maxOffset {
		return 0, io.EOF
	}

	if r.minOffset+int64(len(buf)) > r.maxOffset {
		buf = buf[:r.maxOffset-r.minOffset]
	}

	n, err := r.readerAt.ReadAt(buf, r.minOffset)
	r.minOffset += int64(n)
	return n, err
}
