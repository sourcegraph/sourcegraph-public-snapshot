package types

// A PermissionNamespace represents a distinct context within which permission policies
// are defined and enforced.
type PermissionNamespace string

func (n PermissionNamespace) String() string {
	return string(n)
}

// BatchChangesNamespace represents the Batch Changes namespace.
const BatchChangesNamespace PermissionNamespace = "BATCH_CHANGES"
const SubscriptionsNamespace PermissionNamespace = "SUBSCRIPTIONS"

// Valid checks if a namespace is valid and supported by the Sourcegraph RBAC system.
func (n PermissionNamespace) Valid() bool {
	switch n {
	case BatchChangesNamespace, SubscriptionsNamespace:
		return true
	default:
		return false
	}
}
