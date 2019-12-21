# SAML

Select your SAML identity provider for setup instructions:

- [Okta](/admin/auth/saml_with_okta)
- [Azure Active Directory (Azure AD)](/admin/auth/saml_with_azure_ad)
- [Microsoft Active Directory Federation Services (ADFS)](/admin/auth/saml_with_microsoft_adfs)
- [Auth0](/admin/auth/saml_generic)
- [OneLogin](/admin/auth/saml_generic)
- [Ping Identity](/admin/auth/saml_generic)
- [Salesforce Identity](/admin/auth/saml_generic)
- [Other](/admin/auth/saml_generic)

For advanced SAML configuration options, see the [`saml` auth provider documentation](../config/critical_config.md#saml).

> NOTE: Sourcegraph currently supports at most 1 SAML auth provider at a time (but you can configure additional auth providers of other types). This should not be an issue for 99% of customers.

### SAML troubleshooting

Setting the env var `INSECURE_SAML_LOG_TRACES=1` on the `sourcegraph/server` Docker container (or the `sourcegraph-frontend` pod if Sourcegraph is deployed to a Kubernetes cluster) causes all SAML requests and responses to be logged.
