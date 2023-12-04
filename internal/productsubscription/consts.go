package productsubscription

// AccessTokenPrefix is the prefix used for identifying tokens generated for
// product subscriptions. Currently, this is unused, since we're using
// licensing.LicenseKeyBasedAccessTokenPrefix by default instead, and we retain
// support it for now for back-compat with some existing tokens.
//
// In the future, this prefix can be used if we ever have access tokens for
// product subscriptions that are not based on a Sourcegraph license key.
const AccessTokenPrefix = "sgs_" // "(S)ource(g)raph (S)ubscription"

// GQLErrCodeProductSubscriptionNotFound is the GraphQL error code returned when
// attempting to look up a product subscription failed by any means.
const GQLErrCodeProductSubscriptionNotFound = "ErrProductSubscriptionNotFound"
