package conf

// AuthProvider returns the auth.provider value, applying the default "builtin" if none is
// specified, and respecting the legacy non-auth.provider configs (i.e., setting "samlSPCert" will
// enable the SAML auth provider even if auth.provider is not set to "saml").
func AuthProvider() string {
	c := Get()
	switch {
	case c.AuthProvider == "openidconnect" || authOpenIDConnect(c) != nil:
		return "openidconnect"
	case c.AuthProvider == "saml" || authSAML(c) != nil:
		return "saml"
	case c.AuthProvider == "http-header" || authHTTPHeader(c) != "":
		return "http-header"
	default:
		return "builtin"
	}
}
