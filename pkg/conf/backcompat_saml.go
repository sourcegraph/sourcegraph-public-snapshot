package conf

import "github.com/sourcegraph/sourcegraph/schema"

// authSAML returns the SAMLAuthProvider from the config (if enabled), regardless of whether the old
// saml* or new auth.saml properties were used.
func authSAML(input *schema.SiteConfiguration) (p *schema.SAMLAuthProvider) {
	// oidc* properties first (lower precedence)
	if input.SamlIDProviderMetadataURL != "" || input.SamlSPCert != "" || input.SamlSPKey != "" {
		p = &schema.SAMLAuthProvider{
			IdentityProviderMetadataURL: input.SamlIDProviderMetadataURL,
			ServiceProviderCertificate:  input.SamlSPCert,
			ServiceProviderPrivateKey:   input.SamlSPKey,
		}
	}

	// auth.saml next (higher precedence)
	if input.AuthProvider == "saml" && input.AuthSaml != nil {
		p = input.AuthSaml
	}

	if p != nil {
		p.Type = "saml"
	}

	return p
}
