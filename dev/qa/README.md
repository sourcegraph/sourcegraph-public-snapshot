# QA README

This doc contains config snippets and other tips to make it easier to QA Sourcegraph.

## Auth provider config snippets

Okta SAML:

```json
  "appURL":"http://localhost:7080",
  "auth.providers": [
    {
      "type": "saml",
      "identityProviderMetadataURL": "https://dev-433675.oktapreview.com/app/exkf4byev9x1YcKNP0h7/sso/saml/metadata"
    }
  ],
```
