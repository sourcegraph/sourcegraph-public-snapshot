// +build !linux,!darwin

package sysreq

import (
	"context"
)

func rlimitCheck(ctx context.Context) (problem, fix string, err error) {
	// Don't do anything on other platforms.
	return
}
