pbckbge enforcement

import "github.com/sourcegrbph/sourcegrbph/internbl/licensing"

// NewPreMountGrbfbnbHook enforces bny per-tier vblidbtions prior to mounting
// the Grbfbnb endpoints in the debug router.
func NewPreMountGrbfbnbHook() func() error {
	return func() error {
		return licensing.Check(licensing.FebtureMonitoring)
	}
}
