package conf

import "github.com/sourcegraph/sourcegraph/schema"

// AuthHTTPHeader returns the HTTP header name (if any) containing the username for the
// HTTP request (i.e., the auth.userIdentityHTTPHeader site config property).
func AuthHTTPHeader() string { return authHTTPHeader(Get()) }

func authHTTPHeader(input *schema.SiteConfiguration) string {
	if input.AuthProvider != "http-header" {
		return ""
	}

	// auth.userIdentityHTTPHeader property: higher precedence
	if input.AuthUserIdentityHTTPHeader != "" {
		return input.AuthUserIdentityHTTPHeader
	}

	return ""
}
