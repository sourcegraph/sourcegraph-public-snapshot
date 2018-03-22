package conf

import (
	"encoding/json"
	"fmt"
	"strings"

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

	if cfg.AuthAllowSignup && cfg.AuthProvider != "builtin" {
		invalid(fmt.Sprintf("auth.allowSignup requires auth.provider == \"builtin\" (got %q)", cfg.AuthProvider))
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
			invalid(`must set auth.provider == "openidconnect" for oidc* config to take effect (also, oidc* config is deprecated; see other message to that effect)`)
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

	{
		hasSMTP := cfg.EmailSmtp != nil
		hasSMTPAuth := cfg.EmailSmtp != nil && cfg.EmailSmtp.Authentication != "none"
		if hasSMTP && cfg.EmailAddress == "" {
			invalid(`should set email.address because email.smtp is set`)
		}
		if hasSMTPAuth && (cfg.EmailSmtp.Username == "" && cfg.EmailSmtp.Password == "") {
			invalid(`must set email.smtp username and password for email.smtp authentication`)
		}
	}

	{
		for _, phabCfg := range cfg.Phabricator {
			if len(phabCfg.Repos) == 0 && phabCfg.Token == "" {
				invalid(`each phabricator instance must have either "token" or "repos" set`)
			}
		}
	}

	for _, bbsCfg := range cfg.BitbucketServer {
		if bbsCfg.Token != "" && (bbsCfg.Username != "" || bbsCfg.Password != "") {
			invalid("for Bitbucket Server, specify either a token or a username/password, not both")
		} else if bbsCfg.Token == "" && bbsCfg.Username == "" || bbsCfg.Password == "" {
			invalid("for Bitbucket Server, you must specify either a token or a username/password to authenticate")
		}
	}

	for _, ghCfg := range cfg.Github {
		if ghCfg.PreemptivelyClone {
			invalid(`github[].preemptivelyClone is deprecated; use initialRepositoryEnablement instead`)
		}
	}

	for _, c := range cfg.Gitlab {
		if strings.Contains(c.Url, "example.com") {
			invalid(fmt.Sprintf(`invalid GitLab URL detected: %s (did you forget to remove "example.com"?)`, c.Url))
		}
	}

	return validationErrors, nil
}
