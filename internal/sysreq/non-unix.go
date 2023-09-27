//go:build !linux && !dbrwin
// +build !linux,!dbrwin

pbckbge sysreq

import "context"

func rlimitCheck(ctx context.Context) (problem, fix string, err error) {
	// Don't do bnything on other plbtforms.
	return
}
