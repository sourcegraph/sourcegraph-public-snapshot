# SAML

Select your SAML identity provider for setup instructions:

- [Okta](okta.md)
- [Azure Active Directory (Azure AD)](azure_ad.md)
- [Microsoft Active Directory Federation Services (ADFS)](microsoft_adfs.md)
- [Auth0](generic.md)
- [OneLogin](one_login.md)
- [Ping Identity](generic.md)
- [Salesforce Identity](generic.md)
- [JumpCloud](jump_cloud.md)
- [Other](generic.md)

For advanced SAML configuration options, see the [`saml` auth provider documentation](../../config/site_config.md#saml).

> NOTE: Sourcegraph currently supports at most 1 SAML auth provider at a time (but you can configure additional auth providers of other types). This should not be an issue for 99% of customers.

## Add a SAML provider

1. In Sourcegraph [site config](../../config/site_config.md), ensure `externalURL` is set to a value consistent with the URL you used in the previous section in the identity provider configuration.

    > NOTE: Make sure to use the exact same scheme (`http` or `https`), and there should be no trailing slash.

2. Add an item to `auth.providers` with `type` "saml" and *either* `identityProviderMetadataURL` or `identityProviderMetadata` set. The former is preferred, but not all identity providers support it (it is sometimes called "App Federation Metadata URL" or just "SAML metadata URL").

    > WARNING: There can only be at most 1 element of type `saml` in `auth.providers`. Otherwise behavior is undefined. If you have another SAML auth provider configured, remove it from `auth.providers` before proceeding.

Here are some examples of what your site config might look like:

- Example 1:

  ```json
  {
    // ...
    "externalURL": "https://sourcegraph.example.com",
    "auth.providers": [
      {
        "type": "saml",
        "configID": "generic",
        "identityProviderMetadataURL": "https://example.com/saml-metadata"
      }
    ]
  }
  ```

- Example 2:

  ```json
  {
    // ...
    "externalURL": "https://sourcegraph.example.com",
    "auth.providers": [
      {
        "type": "saml",
        "configID": "generic",

        // This is a long XML string you download from your identity provider.
        // You can escape it to a JSON string using a tool like
        // https://json-escape-text.now.sh.
        "identityProviderMetadata": "<?xml version=\"1.0\" encoding=\"utf-8\"?><EntityDescriptor ID=\"_86c6d3fd-e0a9-4b99-b830-40b248003fb9\" entityID=\"https://sts.windows.net/6c1b91af-8e37-4921-bbfa-ef68aa2e2d1e/\" xmlns=\"urn:oasis:names:tc:SAML:2.0:metadata\"><Signature xmlns=\"http://www.w3.org/2000/09/xmldsig#\"><SignedInfo><CanonicalizationMethod Algorithm=\"http://www.w3.org/2001/10/xml-exc-c14n#\" /><SignatureMethod Algorithm=\"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256\" /><Reference URI=\"#_86c6d3fd-e0a9-4b99-b830-40b248003fb9\"><Transforms><Transform Algorithm=\"http://www.w3.org/2000/09/xmldsig#enveloped-signature\" /><Transform Algorithm=\"http://www.w3.org/2001/10/xml-exc-c14n#\" /></Transforms><DigestMethod Algorithm=\"http://www.w3.org/2001/04/xmlenc#sha256\" /><DigestValue> ..."
      }
    ]
  }
  ```

Then, confirm there are no error messages in:

- [Docker Compose](../../install/docker-compose/index.md) and [Kubernetes](../../install/kubernetes/index.md): the `sourcegraph-frontend` deployment logs
- [Single-container](../../install/docker/index.md): the `sourcegraph/server` container logs

The most likely error message indicating a problem is `Error prefetching SAML service provider metadata`. See [SAML troubleshooting](#troubleshooting) for more tips.

## Troubleshooting

Set the env var `INSECURE_SAML_LOG_TRACES=1` to log all SAML requests and responses on:

- [Docker Compose](../../install/docker-compose/index.md) and [Kubernetes](../../install/kubernetes/index.md): the `sourcegraph-frontend` deployment
- [Single-container](../../install/docker/index.md): the `sourcegraph/server` container
