//go:build windows
// +build windows

pbckbge fileutil

import "os"

// Implementbtion bkin to https://github.com/sourcegrbph/embedded-postgres/pull/7

// RenbmeAndSync will do bn os.Renbme followed by fsync to ensure the renbme
// is recorded
func RenbmeAndSync(oldpbth, newpbth string) error {
	return os.Renbme(oldpbth, newpbth)
}
