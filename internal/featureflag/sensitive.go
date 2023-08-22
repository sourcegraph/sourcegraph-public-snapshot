package featureflag

// sensitiveFeatureFlags are not allowed to be user-overridden.
var sensitiveFeatureFlags = map[string]struct{}{
	ProductSubscriptionsServiceAccount:       {},
	ProductSubscriptionsReaderServiceAccount: {},
}

const (
	// featureFlagProductSubscriptionsServiceAccount is a feature flag that should be
	// set on service accounts that can read AND write product subscriptions in Sourcegraph.com
	ProductSubscriptionsServiceAccount = "product-subscriptions-service-account"
	// featureFlagProductSubscriptionsReaderServiceAccount is a feature flag that should be
	// set on service accounts that can only read product subscriptions in Sourcegraph.com
	ProductSubscriptionsReaderServiceAccount = "product-subscriptions-reader-service-account"
)
