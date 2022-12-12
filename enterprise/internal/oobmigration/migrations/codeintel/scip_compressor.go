package codeintel

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var compressor = &gzipCompressor{
	writers: sync.Pool{
		New: func() any { return gzip.NewWriter(nil) },
	},
}

type gzipCompressor struct {
	writers sync.Pool
}

func (c *gzipCompressor) compress(r io.Reader) ([]byte, error) {
	gzipWriter := c.writers.Get().(*gzip.Writer)
	defer c.writers.Put(gzipWriter)
	compressBuf := new(bytes.Buffer)
	gzipWriter.Reset(compressBuf)

	if _, err := io.Copy(gzipWriter, r); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}

	return compressBuf.Bytes(), nil
}
