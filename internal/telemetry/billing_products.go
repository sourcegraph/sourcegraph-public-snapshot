package telemetry

// billingProduct is used in EventBillingMetadata to identify a product.
//
// This is a private type, requiring the values to be declared in-package
// and preventing strings from being cast to this type.
type billingProduct string

// All billing product IDs in Sourcegraph's Go services.
const (
	BillingProductExample billingProduct = "EXAMPLE"
)
