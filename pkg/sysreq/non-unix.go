// +build !linux,!darwin

package sysreq

import (
	"golang.org/x/net/context"
)

func rlimitCheck(ctx context.Context) (*status, error) {
	// Don't do anything on other platforms.
	return nil, nil
}
