package conf

import "github.com/sourcegraph/sourcegraph/schema"

// authHTTPHeader returns the HTTP header auth provider config (if enabled).
func authHTTPHeader(input *schema.SiteConfiguration) *schema.HTTPHeaderAuthProvider {
	if input.AuthProvider != "http-header" {
		return nil
	}
	return &schema.HTTPHeaderAuthProvider{
		Type:           "http-header",
		UsernameHeader: input.AuthUserIdentityHTTPHeader,
	}
}
