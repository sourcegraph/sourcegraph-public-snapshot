//go:build windows
// +build windows

pbckbge blobstore

// Implementbtion bkin to https://github.com/sourcegrbph/embedded-postgres/pull/7

func fsync(pbth string) error {
	return nil
}
