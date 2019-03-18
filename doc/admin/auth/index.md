# User authentication (SSO)

> NOTE: Single sign-on integrations are a [paid upgrade](https://about.sourcegraph.com/pricing).

Sourcegraph supports the following ways for users to sign in:

- [Builtin](#builtin-authentication)
- [OpenID Connect](#openid-connect) (including [Google accounts on G Suite](#g-suite-google-accounts))
- [SAML](#saml)
- [HTTP authentication proxies](#http-authentication-proxies)

The authentication provider is configured in the [`auth.providers`](../site_config/index.md#code-classlanguage-textauthproviders-array) site configuration option.

## Builtin authentication

The [`builtin` auth provider](../site_config/index.md#code-classlanguage-textbuiltinauthprovider-object) manages user accounts internally in its own database. It supports user signup, login, and password reset (via email if configured, or else via a site admin).

Site configuration example:

```json
{
  // ...,
  "auth.providers": [{ "type": "builtin", "allowSignup": true }]
}
```

The top-level [`auth.public`](../site_config/index.md#code-classlanguage-textauthpublic-boolean) (default `false`) site configuration option controls whether anonymous users are allowed to access and use the site without being signed in .

## OpenID Connect

The [`openidconnect` auth provider](../site_config/index.md#code-classlanguage-textopenidconnectauthprovider-object) authenticates users via OpenID Connect, which is supported by many external services, including:

- [G Suite (Google accounts)](#g-suite-google-accounts)
- [Okta](https://developer.okta.com/docs/api/resources/oidc.html)
- [Ping Identity](https://www.pingidentity.com/en/resources/client-library/articles/openid-connect.html)
- [Auth0](https://auth0.com/docs/protocols/oidc)
- [Salesforce Identity](https://developer.salesforce.com/page/Inside_OpenID_Connect_on_Force.com)
- [Microsoft Azure Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-protocols-openid-connect-code)
- [Google Identity Platform](https://developers.google.com/identity/)
- Known issue: [OneLogin](https://www.onelogin.com/openid-connect) OpenID Connect is not supported (use SAML for OneLogin instead)

To configure Sourcegraph to authenticate users via OpenID Connect:

1.  Create a new OpenID Connect client in the external service (such as one of those listed above).
    - **Redirect/callback URI:** `https://sourcegraph.example.com/.auth/callback` (replace `https://sourcegraph.example.com` with the value of the `appURL` property in your config)
1.  Provide the OpenID Connect client's issuer, client ID, and client secret in the Sourcegraph site configuration shown below.
1.  (Optional) Require users to have a specific email domain name to authenticate (e.g., to limit users to only those from your organization).

Example [`openidconnect` auth provider](../site_config/index.md#code-classlanguage-textopenidconnectauthprovider-object) configuration:

```json
{
  // ...
  "appURL": "https://sourcegraph.example.com",
  "auth.providers": [
    {
      "type": "openidconnect",
      "issuer": "https://oidc.example.com",
      "clientID": "my-client-id",
      "clientSecret": "my-client-secret",
      "requireEmailDomain": "example.com"
    }
  ]
}
```

Sourcegraph supports the OpenID Connect Discovery standard for configuring the auth provider (using the document at, e.g., `https://oidc.example.com/.well-known/openid-configuration`).

See the [`openid` auth provider documentation](../site_config/index.md#code-classlanguage-textopenidconnectauthprovider-object) for the full set of configuration options.

### G Suite (Google accounts)

Google's G Suite supports OpenID Connect, which is the best way to enable Sourcegraph authentication using Google accounts. To set it up:

1.  Create an **OAuth client ID** and client secret in the [Google API credentials console](https://console.developers.google.com/apis/credentials). [Google's interactive OpenID Connect documentation page](https://developers.google.com/identity/protocols/OpenIDConnect#getcredentials):
    - **Application type:** Web application
    - **Name:** Sourcegraph (or any other name your users will recognize)
    - **Authorized JavaScript origins:** (leave blank)
    - **Authorized redirect URIs:** `https://sourcegraph.example.com/.auth/callback` (replace `https://sourcegraph.example.com` with the value of the `appURL` property in your config)
1.  Use the **client ID** and **client secret** values in Sourcegraph site configuration (as shown in the example below).
1.  Set your G Suite domain in `requireEmailDomain` to prevent users outside your organization from signing in.

Example [`openidconnect` auth provider](../site_config/index.md#code-classlanguage-textopenidconnectauthprovider-object) configuration for G Suite:

```json
{
  // ...
  "appURL": "https://sourcegraph.example.com",
  "auth.providers": [
    {
      "type": "openidconnect",
      "issuer": "https://accounts.google.com", // All G Suite domains use this issuer URI.
      "clientID": "my-client-id",
      "clientSecret": "my-client-secret",
      "requireEmailDomain": "example.com"
    }
  ]
}
```

## SAML

The [`saml` auth provider](../site_config/index.md#code-classlanguage-textsamlauthprovider-object) authenticates users via SAML 2.0, which is supported by many external services, including:

- [Okta](https://developer.okta.com/standards/SAML/index) - _Admin > Classic UI > Applications > Add Application > Create New App > Web / SAML 2.0_
- [OneLogin](https://www.onelogin.com/saml) - _Administration > Apps > Add Apps > SAML Test Connector (SP)_
- [Ping Identity](https://www.pingidentity.com/en/resources/client-library/articles/saml.html)
- [Auth0](https://auth0.com/docs/protocols/saml)
- [Salesforce Identity](https://help.salesforce.com/articleView?id=sso_saml_setting_up.htm&type=0)
- [Microsoft Azure Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/develop/active-directory-single-sign-on-protocol-reference)

To configure Sourcegraph to authenticate users via SAML:

1.  Register Sourcegraph as a SAML Service Provider in the external SAML Identity Provider (such as one of those listed above). Use the following settings (the exact names and labels vary across services).
    - **Assertion Consumer Service URL, Recipient URL, Destination URL, Single sign-on URL:** `https://sourcegraph.example.com/.auth/saml/acs` (substituting your [`appURL`](../site_config/index.md#appurl-string))
    - **Service Provider (issuer, entity ID, audience URI, metadata URL):** `https://sourcegraph.example.com/.auth/saml/metadata` (substituting your [`appURL`](../site_config/index.md#appurl-string))
    - **Attribute statements:**
      - `email` (required): the user's email
      - `login` (optional): the user's username
      - `displayName` (optional): the full name of the user
1.  Obtain the SAML identity provider metadata URL and use it in Sourcegraph site configuration as shown below.

Example [`saml` auth provider](../site_config/index.md#code-classlanguage-textsamlauthprovider-object) configuration:

```json
{
  // ...
  "appURL": "https://sourcegraph.example.com",
  "auth.providers": [
    {
      "type": "saml",
      // Find this listed in your SAML Identity Provider's admin interface after you've added your SAML app.
      // It's sometimes called the "Identity Provider metadata URL" or "SAML metadata".
      "identityProviderMetadataURL": "https://example.com/saml-metadata"
    }
  ]
}
```

See the [`saml` auth provider documentation](../site_config/index.md#code-classlanguage-textsamlauthprovider-object) for the full set of configuration options.

### SAML troubleshooting

Setting the env var `INSECURE_SAML_LOG_TRACES=1` on the `sourcegraph/server` Docker conatiner (or the `sourcegraph-frontend` pod if Sourcegraph is deployed to a Kubernetes cluster) causes all SAML requests and responses to be logged.

### Vendor-specific SAML instructions

- [Configuring SAML with OneLogin](saml_with_onelogin.md)
- [Configuring SAML with Microsoft Active Directory Federation Services (ADFS)](saml_with_microsoft_adfs.md)

### SAML Service Provider metadata

To make it easier to register Sourcegraph as a SAML Service Provider with your SAML Identity Provider, Sourcegraph exposes SAML Service Provider metadata at the URL path `/.auth/saml/metadata`. For example:

```
https://sourcegraph.example.com/.auth/saml/metadata
```

## HTTP authentication proxies

You can wrap Sourcegraph in an authentication proxy that authenticates the user and passes the user's username to Sourcegraph via HTTP headers. The most popular such authentication proxy is [bitly/oauth2_proxy](https://github.com/bitly/oauth2_proxy), which works well with Sourcegraph.

To use an authentication proxy to authenticate users to Sourcegraph, add the following lines to your site configuration:

```json
{
  // ...
  "auth.providers": [
    {
      "type": "http-header",
      "usernameHeader": "X-Forwarded-User"
    }
  ]
}
```

Replace `X-Forwarded-User` with the name of the HTTP header added by the authentication proxy that contains the user's username.

Ensure that the HTTP proxy is not setting its own `Authorization` header on the request. Sourcegraph rejects requests with unrecognized `Authorization` headers and prints the error log `lvl=eror msg="Invalid Authorization header." err="unrecognized HTTP Authorization request header scheme (supported values: token, token-sudo)"`.

For bitly/oauth2_proxy, use the `-pass-basic-auth false` option to prevent it from sending the `Authorization` header.

## Username normalization

Usernames on Sourcegraph are normalized according to the following rules.

- Any portion of the username after a '@' character is removed
- Any characters not in `[a-zA-Z0-9-]` are replaced with `-`
- Usernames with consecutive '-' characters are not allowed
- Usernames that start or end with '-' are not allowed

Usernames from authentication providers are normalized before being used in Sourcegraph. Usernames chosen by users are rejected if they do not meet these criteria.

For example, a user whose external username (according the authentication provider) is `alice.smith@example.com` would have the Sourcegraph username `alice-smith`.

If multiple accounts normalize into the same username, only the first user account is created. Other users won't be able to sign in. This is a rare occurrence; contact support if this is a blocker.
