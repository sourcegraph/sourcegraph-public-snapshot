package tenant

import "github.com/sourcegraph/sourcegraph/internal/env"

var enforcementMode = env.Get("SRC_TENANT_ENFORCEMENT_MODE", "disabled", "INTERNAL: enforcement mode for tenant isolation. Valid values: disabled, logging, strict")

// ShouldLogNoTenant returns true if the tenant enforcement mode is logging or strict.
// It is used to log a warning if a request to a low-level store is made without a tenant
// so we can identify missing tenants. This will go away and only strict will be allowed
// once we are confident that all contexts carry tenants.
func ShouldLogNoTenant() bool {
	switch enforcementMode {
	case "logging", "strict":
		return true
	default:
		return false
	}
}

// EnforceTenant returns true if the tenant enforcement mode is strict.
// Stores should use this to enforce tenant isolation.
// This will go away and only strict will be allowed once we are confident that
// isolation is working and this is not breaking existing instances.
func EnforceTenant() bool {
	switch enforcementMode {
	case "strict":
		return true
	default:
		return false
	}
}
