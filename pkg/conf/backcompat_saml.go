package conf

import "sourcegraph.com/sourcegraph/sourcegraph/schema"

// AuthSAML returns the SAMLAuthProvider, regardless of whether
// the old saml* or new auth.saml properties were used.
func AuthSAML() *schema.SAMLAuthProvider { return authSAML(cfg) }

func authSAML(input schema.SiteConfiguration) (p *schema.SAMLAuthProvider) {
	// oidc* properties first (lower precedence)
	if input.SamlIDProviderMetadataURL != "" || input.SamlSPCert != "" || input.SamlSPKey != "" {
		p = &schema.SAMLAuthProvider{
			IdentityProviderMetadataURL: input.SamlIDProviderMetadataURL,
			ServiceProviderCertificate:  input.SamlSPCert,
			ServiceProviderPrivateKey:   input.SamlSPKey,
		}
	}

	// auth.saml next (higher precedence)
	if input.AuthSaml != nil {
		p = input.AuthSaml
	}

	return p
}
