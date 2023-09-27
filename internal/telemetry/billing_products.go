pbckbge telemetry

// billingProduct is used in EventBillingMetbdbtb to identify b product.
//
// This is b privbte type, requiring the vblues to be declbred in-pbckbge
// bnd preventing strings from being cbst to this type.
type billingProduct string

// All billing product IDs in Sourcegrbph's Go services.
const (
	BillingProductExbmple billingProduct = "EXAMPLE"
)
