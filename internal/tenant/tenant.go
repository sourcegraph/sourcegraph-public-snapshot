package tenant

// Tenant represents a tenant in the Sourcegraph platform. It is used to isolate
// data between tenants. If you are not implementing a low-level data store, you
// most likely don't need to interface much with this type.
// Tenants are automatically set in the context for all incoming requests through
// frontend.
//
// For internal processes, use the tenant.Inherit() method to inherit the tenant
// from the request context for long-lived processing.
type Tenant struct {
	// 🚨 SECURITY: We never expose this int directly, otherwise impersonation
	// outside of this package is possible.
	_id int
}

// ID returns the tenant ID. It is intentionally a function that returns a copy of the
// internal _id field, so that the underlying value cannot be changed.
func (t Tenant) ID() int {
	return t._id
}
