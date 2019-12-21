# SAML

Select your SAML identity provider for setup instructions:

- [Okta](/admin/auth/saml_with_okta)
- [OneLogin](/admin/auth/saml_with_onelogin)
- [Azure Active Directory (Azure AD)](/admin/auth/saml_with_azure_ad)
- [Microsoft Active Directory Federation Services (ADFS)](/admin/auth/saml_with_microsoft_adfs)
- [Auth0](/admin/auth/saml_generic)
- [Ping Identity](/admin/auth/saml_generic)
- [Salesforce Identity](/admin/auth/saml_generic)
- [Other](/admin/auth/saml_generic)

For advanced SAML configuration options, see the [`saml` auth provider documentation](../config/critical_config.md#saml).

<!--

The [`saml` auth provider](../config/critical_config.md#saml) authenticates users via SAML 2.0, which is supported by many external services, including:

- [Okta](https://developer.okta.com/standards/SAML/index) - _Admin > Classic UI > Applications > Add Application > Create New App > Web / SAML 2.0_
- [OneLogin](https://www.onelogin.com/saml) - _Administration > Apps > Add Apps > SAML Test Connector (SP)_
- [Ping Identity](https://www.pingidentity.com/en/resources/client-library/articles/saml.html)
- [Auth0](https://auth0.com/docs/protocols/saml)
- [Salesforce Identity](https://help.salesforce.com/articleView?id=sso_saml_setting_up.htm&type=0)
- [Microsoft Azure Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-single-sign-on-protocol-reference)
-->


> NOTE: Sourcegraph currently supports at most 1 SAML auth provider at a time (but you can configure additional auth providers of other types). This should not be an issue for 99% of customers.

### SAML troubleshooting

Setting the env var `INSECURE_SAML_LOG_TRACES=1` on the `sourcegraph/server` Docker container (or the `sourcegraph-frontend` pod if Sourcegraph is deployed to a Kubernetes cluster) causes all SAML requests and responses to be logged.
