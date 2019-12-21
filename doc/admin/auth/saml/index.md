# SAML

Select your SAML identity provider for setup instructions:

- [Okta](/admin/auth/saml/okta)
- [Azure Active Directory (Azure AD)](/admin/auth/saml/azure_ad)
- [Microsoft Active Directory Federation Services (ADFS)](/admin/auth/saml/microsoft_adfs)
- [Auth0](/admin/auth/saml/generic)
- [OneLogin](/admin/auth/saml/generic)
- [Ping Identity](/admin/auth/saml/generic)
- [Salesforce Identity](/admin/auth/saml/generic)
- [Other](/admin/auth/saml/generic)

For advanced SAML configuration options, see the [`saml` auth provider documentation](../config/critical_config.md#saml).

> NOTE: Sourcegraph currently supports at most 1 SAML auth provider at a time (but you can configure additional auth providers of other types). This should not be an issue for 99% of customers.

### SAML troubleshooting

Setting the env var `INSECURE_SAML_LOG_TRACES=1` on the `sourcegraph/server` Docker container (or the `sourcegraph-frontend` pod if Sourcegraph is deployed to a Kubernetes cluster) causes all SAML requests and responses to be logged.
