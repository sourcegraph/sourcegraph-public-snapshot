package enforcement

import "github.com/sourcegraph/sourcegraph/internal/licensing"

// PreMountGrafanaHook enforces any per-tier validations prior to mounting
// the Grafana endpoints in the debug router.
func PreMountGrafanaHook() error {
	return licensing.Check(licensing.FeatureMonitoring)
}
