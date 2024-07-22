package limitedgzip

import (
	"compress/gzip"
	"io"
)

// WithReader returns a new io.ReadCloser that reads and decompresses the body
// it reads until io.EOF or the specified limit is reached.
func WithReader(body io.ReadCloser, limit int64) (io.ReadCloser, error) {
	gzipReader, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}

	body = struct {
		io.Reader
		io.Closer
	}{
		Reader: io.LimitReader(gzipReader, limit),
		Closer: gzipReader,
	}

	return body, nil
}
