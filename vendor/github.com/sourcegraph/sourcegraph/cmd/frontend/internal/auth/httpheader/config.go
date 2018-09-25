package httpheader

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// getProviderConfig returns the HTTP header auth provider config. At most 1 can be specified in
// site config; if there is more than 1, it returns multiple == true (which the caller should handle
// by returning an error and refusing to proceed with auth).
func getProviderConfig() (pc *schema.HTTPHeaderAuthProvider, multiple bool) {
	for _, p := range conf.AuthProviders() {
		if p.HttpHeader != nil {
			if pc != nil {
				return pc, true // multiple http-header auth providers
			}
			pc = p.HttpHeader
		}
	}
	return pc, false
}

func validateConfig(c *schema.SiteConfiguration) (problems []string) {
	var httpHeaderAuthProviders int
	for _, p := range conf.AuthProvidersFromConfig(c) {
		if p.HttpHeader != nil {
			httpHeaderAuthProviders++
		}
	}
	if httpHeaderAuthProviders >= 2 {
		problems = append(problems, `at most 1 http-header auth provider may be used`)
	}

	hasSingularAuthHTTPHeader := c.AuthUserIdentityHTTPHeader != ""
	if c.AuthProvider == "http-header" && !hasSingularAuthHTTPHeader {
		problems = append(problems, `auth.userIdentityHTTPHeader must be configured when auth.provider == "http-header"`)
	}
	if hasSingularAuthHTTPHeader {
		problems = append(problems, `auth.userIdentityHTTPHeader is deprecated; use "auth.providers" with an entry of {"type":"http-header","usernameHeader":"..."} instead`)
		if c.AuthProvider != "http-header" {
			problems = append(problems, `must set auth.provider == "http-header" for auth.userIdentityHTTPHeader config to take effect`)
		}
	}
	return problems
}
