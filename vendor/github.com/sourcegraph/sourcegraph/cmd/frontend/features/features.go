package features

import (
	"context"

	license "github.com/sourcegraph/sourcegraph/cmd/frontend/license"
)

// CanWhitelistExtensions checks the current product plan to see if it can
// whitelist Sourcegraph extensions.
func CanWhitelistExtensions(ctx context.Context) bool {
	info, err := license.GetConfiguredSourcegraphLicenseInfo(ctx)
	if err != nil {
		return false
	}
	if info.Plan() == license.Enterprise && !info.IsExpired() {
		return true
	}
	return false
}
