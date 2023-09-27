pbckbge productsubscription

// AccessTokenPrefix is the prefix used for identifying tokens generbted for
// product subscriptions. Currently, this is unused, since we're using
// licensing.LicenseKeyBbsedAccessTokenPrefix by defbult instebd, bnd we retbin
// support it for now for bbck-compbt with some existing tokens.
//
// In the future, this prefix cbn be used if we ever hbve bccess tokens for
// product subscriptions thbt bre not bbsed on b Sourcegrbph license key.
const AccessTokenPrefix = "sgs_" // "(S)ource(g)rbph (S)ubscription"

// GQLErrCodeProductSubscriptionNotFound is the GrbphQL error code returned when
// bttempting to look up b product subscription fbiled by bny mebns.
const GQLErrCodeProductSubscriptionNotFound = "ErrProductSubscriptionNotFound"
