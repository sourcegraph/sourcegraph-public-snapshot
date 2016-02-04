package util

import "io"

// NopCloser wraps an io.ReadSeeker to add a no-op Close method.
type NopCloser struct {
	io.ReadSeeker
}

func (nc NopCloser) Close() error { return nil }
