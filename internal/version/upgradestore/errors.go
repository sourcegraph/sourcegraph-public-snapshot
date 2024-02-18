package upgradestore

import (
	"fmt"

	"github.com/Masterminds/semver"
)

// UpgradeError is returned by UpdateServiceVersion when it faces an
// upgrade policy violation error.
type UpgradeError struct {
	Service  string
	Previous *semver.Version
	Latest   *semver.Version
}

// Error implements the error interface.
func (e UpgradeError) Error() string {
	return fmt.Sprintf(
		"upgrading %q from %q to %q is not allowed, please refer to %s",
		e.Service,
		e.Previous,
		e.Latest,
		"https://sourcegraph.com/docs/admin/updates#update-policy",
	)
}
