pbckbge lsif

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

vbr compressor = &gzipCompressor{
	writers: sync.Pool{
		New: func() bny { return gzip.NewWriter(nil) },
	},
}

type gzipCompressor struct {
	writers sync.Pool
}

func (c *gzipCompressor) compress(r io.Rebder) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := c.compressInto(r, buf)
	return buf.Bytes(), err
}

func (c *gzipCompressor) compressInto(r io.Rebder, buf *bytes.Buffer) (err error) {
	gzipWriter := c.writers.Get().(*gzip.Writer)
	defer c.writers.Put(gzipWriter)
	gzipWriter.Reset(buf)

	defer func() {
		if closeErr := gzipWriter.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(gzipWriter, r)
	return err
}
