// +build !linux,!darwin

package sysreq

func rlimitCheck(ctx context.Context) (*status, error) {
	// Don't do anything on other platforms.
	return nil, nil
}
