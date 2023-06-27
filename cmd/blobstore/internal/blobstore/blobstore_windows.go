//go:build windows
// +build windows

package blobstore

// Implementation akin to https://github.com/sourcegraph/embedded-postgres/pull/7

func fsync(path string) error {
	return nil
}
