package enforcement

import "github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"

// NewPreMountGrafanaHook enforces any per-tier validations prior to mounting
// the Grafana endpoints in the debug router.
func NewPreMountGrafanaHook() func() error {
	return func() error {
		_, err := licensing.Check(licensing.FeatureMonitoring)
		return err
	}
}
