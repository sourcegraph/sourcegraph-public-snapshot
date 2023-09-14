package uploadstore

import "io"

type closeWrapper struct {
	io.ReadCloser
	close func()
}

func (c *closeWrapper) Close() error {
	c.ReadCloser.Close()
	c.close()
	return nil
}

// NewExtraCloser returns wraps a ReadCloser with an extra close function
// that will be called after the underlying ReadCloser has been closed.
func NewExtraCloser(rc io.ReadCloser, close func()) io.ReadCloser {
	return &closeWrapper{ReadCloser: rc, close: close}
}
