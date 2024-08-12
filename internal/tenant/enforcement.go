package tenant

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

var enforcementMode = env.Get("SRC_TENANT_ENFORCEMENT_MODE", "disabled", "INTERNAL: enforcement mode for tenant isolation. Valid values: disabled, logging, strict")

// shouldLogNoTenant returns true if the tenant enforcement mode is logging or strict.
// It is used to log a warning if a request to a low-level store is made without a tenant
// so we can identify missing tenants. This will go away and only strict will be allowed
// once we are confident that all contexts carry tenants.
func shouldLogNoTenant() bool {
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

// MockEnforceTenant sets the tenant enforcement mode to strict for the duration
// of the test. It automatically resets the enforcement mode to its previous
// value after the test.
//
// Note: You must not call this function outside of tests, it will panic.
func MockEnforceTenant(t *testing.T) {
	if !testutil.IsTest {
		panic("only call this function in tests")
	}
	old := enforcementMode
	enforcementMode = "strict"
	t.Cleanup(func() { enforcementMode = old })
}
