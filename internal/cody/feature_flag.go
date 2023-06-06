package cody

import "context"

// IsCodyEnabled is the OSS Cody check. This function gets
// overridden in enterprise code.
var IsCodyEnabled = func(ctx context.Context) bool {
	return false
}
