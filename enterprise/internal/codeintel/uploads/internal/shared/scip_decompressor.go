package shared

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var Decompressor = &gzipDecompressor{
	readers: sync.Pool{
		New: func() any { return new(gzip.Reader) },
	},
}

type gzipDecompressor struct {
	readers sync.Pool
}

func (c *gzipDecompressor) Decompress(r io.Reader) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := c.decompressInto(r, buf)
	return buf.Bytes(), err
}

func (c *gzipDecompressor) decompressInto(r io.Reader, buf *bytes.Buffer) (err error) {
	gzipReader := c.readers.Get().(*gzip.Reader)
	defer c.readers.Put(gzipReader)

	if err := gzipReader.Reset(r); err != nil {
		return err
	}
	defer func() {
		if closeErr := gzipReader.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(buf, gzipReader)
	return err
}
