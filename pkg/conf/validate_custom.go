package conf

import (
	"encoding/json"
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// ValidateCustom validates the site config using custom validation steps that are not
// able to be expressed in the JSON Schema.
func ValidateCustom(normalizedInput []byte) (validationErrors []string, err error) {
	var cfg schema.SiteConfiguration
	if err := json.Unmarshal(normalizedInput, &cfg); err != nil {
		return nil, err
	}

	invalid := func(msg string) {
		validationErrors = append(validationErrors, msg)
	}

	if cfg.AuthAllowSignup && !(cfg.AuthProvider == "builtin" || cfg.AuthProvider == "auth0") {
		invalid(fmt.Sprintf("auth.allowSignup requires auth.provider == \"builtin\" or \"auth0\" (got %q)", cfg.AuthProvider))
	}

	if cfg.AuthProvider == "openidconnect" && cfg.AppURL == "" {
		invalid(`auth.provider == "openidconnect" requires appURL to be set to the external URL of your site (example: https://sourcegraph.example.com)`)
	}

	{
		hasOldOIDC := cfg.OidcProvider != "" || cfg.OidcClientID != "" || cfg.OidcClientSecret != "" || cfg.OidcEmailDomain != "" || cfg.OidcOverrideToken != ""
		hasNewOIDC := cfg.AuthOpenIDConnect != nil
		if hasOldOIDC && hasNewOIDC {
			invalid(`both oidc* properties and auth.openIDConnect are set; preferring properties from the auth.openIDConnect object (oidc* properties are deprecated)`)
		} else if hasOldOIDC {
			invalid(`oidc* properties are deprecated; use auth.provider == "openidconnect" and the auth.openIDConnect object instead`)
		} else if cfg.AuthProvider == "openidconnect" && !hasOldOIDC && !hasNewOIDC {
			invalid(`auth.openIDConnect must be configured when auth.provider == "openidconnect"`)
		}
		if hasOldOIDC && cfg.AuthProvider != "openidconnect" {
			invalid(`must set auth.provider == "openidconnect" for oidc* config to take effect`)
		}
		if hasNewOIDC && cfg.AuthProvider != "openidconnect" {
			invalid(`must set auth.provider == "openidconnect" for auth.openIDConnect config to take effect`)
		}
	}

	{
		hasOldSAML := cfg.SamlIDProviderMetadataURL != "" || cfg.SamlSPCert != "" || cfg.SamlSPKey != ""
		hasNewSAML := cfg.AuthSaml != nil
		if hasOldSAML && hasNewSAML {
			invalid(`both saml* properties and auth.saml are set; preferring properties from the auth.saml object (saml* properties are deprecated)`)
		} else if hasOldSAML {
			invalid(`saml* properties are deprecated; use auth.provider == "saml" and the auth.saml object instead`)
		} else if cfg.AuthProvider == "saml" && !hasOldSAML && !hasNewSAML {
			invalid(`auth.saml must be configured when auth.provider == "saml"`)
		}
		if hasOldSAML && cfg.AuthProvider != "saml" {
			invalid(`must set auth.provider == "saml" for saml* config to take effect`)
		}
		if hasNewSAML && cfg.AuthProvider != "saml" {
			invalid(`must set auth.provider == "saml" for auth.saml config to take effect`)
		}
	}

	{
		hasOldAuthHTTPHeader := cfg.SsoUserHeader != ""
		hasNewAuthHTTPHeader := cfg.AuthUserIdentityHTTPHeader != ""
		if hasOldAuthHTTPHeader && hasNewAuthHTTPHeader {
			invalid(`both ssoUserHeader and auth.userIdentityHTTPHeader are set; preferring the latter (ssoUserHeader is deprecated)`)
		} else if hasOldAuthHTTPHeader {
			invalid(`ssoUserHeader is deprecated; use auth.provider == "http-header" and the auth.userIdentityHTTPHeader property instead`)
		} else if cfg.AuthProvider == "http-header" && !hasOldAuthHTTPHeader && !hasNewAuthHTTPHeader {
			invalid(`auth.userIdentityHTTPHeader must be configured when auth.provider == "http-header"`)
		}
		if hasOldAuthHTTPHeader && cfg.AuthProvider != "http-header" {
			invalid(`must set auth.provider == "http-header" for ssoUserHeader config to take effect`)
		}
		if hasNewAuthHTTPHeader && cfg.AuthProvider != "http-header" {
			invalid(`must set auth.provider == "http-header" for auth.userIdentityHTTPHeader config to take effect`)
		}
	}

	return validationErrors, nil
}
