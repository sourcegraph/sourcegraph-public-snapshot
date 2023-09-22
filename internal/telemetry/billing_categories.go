package telemetry

// billingCategory is used in EventBillingMetadata to identify a product.
//
// This is a private type, requiring the values to be declared in-package
// and preventing strings from being cast to this type.
type billingCategory string

// All billing category IDs in Sourcegraph's Go services.
const (
	BillingCategoryExample billingCategory = "EXAMPLE"
)
