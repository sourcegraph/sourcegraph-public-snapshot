package conf

import "sourcegraph.com/sourcegraph/sourcegraph/schema"

// AuthHTTPHeader returns the HTTP header name (if any) containing the username for the
// HTTP request, regardless of whether the old ssoUserHeader or new auth.userIdentityHTTPHeader
// property is used.
func AuthHTTPHeader() string { return authHTTPHeader(cfg) }

func authHTTPHeader(input schema.SiteConfiguration) string {
	// auth.userIdentityHTTPHeader property: higher precedence
	if input.AuthUserIdentityHTTPHeader != "" {
		return input.AuthUserIdentityHTTPHeader
	}

	// ssoUserHeader property: lower precedence
	return input.SsoUserHeader
}
