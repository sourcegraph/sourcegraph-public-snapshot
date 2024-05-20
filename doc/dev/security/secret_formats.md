# Sourcegraph Secret Formats

Sourcegraph uses a number of secret formats to store authentication tokens and keys. This page documents each secret type, and the regular expressions that can be used to match each format.

|                  Token Name                  |                                   Description                                    |            Type            |    Regular Expression     |                         |
| -------------------------------------------- | -------------------------------------------------------------------------------- | -------------------------- | ------------------------- | ----------------------- |
| Sourcegraph Access Token (v3)                | Token used to access the Sourcegraph GraphQL API                                 | User-generated             | `sgp_(?:[a-fA-F0-9]{16}\|local)_[a-fA-F0-9]{40}` |
| Sourcegraph Access Token (v2, deprecated)    | Token used to access the Sourcegraph GraphQL API                                 | User-generated             | `sgp_[a-fA-F0-9]{40}`     |                         |
| Sourcegraph Access Token (v1, deprecated)    | Token used to access the Sourcegraph GraphQL API                                 | User-generated             | `[a-fA-F0-9]{40}`         |                         |
| Sourcegraph Dotcom User Gateway Access Token | Token used to grant sourcegraph.com users access to Cody                         | Backend (not user-visible) | `sgd_[a-fA-F0-9]{64}`     |                         |
| Sourcegraph License Key Token                | Token used for product subscriptions, derived from a Sourcegraph license key     | Backend (not user-visible) | `slk_[a-fA-F0-9]{64}`     |                         |
| Sourcegraph Enterprise subscription (aka "product subscription") Token       | Token used for Enterprise subscriptions, derived from a Sourcegraph license key | Backend (not user-visible) | `sgs_[a-fA-F0-9]{64}`     |                         |

For further information about Sourcegraph Access Tokens, see:
- [Creating an access token](https://sourcegraph.com/docs/cli/how-tos/creating_an_access_token)
- [Revoking an access token](https://sourcegraph.com/docs/cli/how-tos/revoking_an_access_token)

Sourcegraph is in the process of rolling out a new Sourcegraph Access Token format. When generating an access token you may receive a token in v2 or v3 format depending on your Sourcegraph instance's version. Newer instances are fully backwards-compatible with all older formats.


### Sourcegraph Access Token (v3) Instance Identifier
The Sourcegraph Access Token (v3) includes an *instance identifier* which can be used by Sourcegraph to securely identify which instance the token was generated for. In the event of a token leak, this allows us to inform the relevant customer.

```
sgp _ <instance identifier> _ <token value>
```

The *instance identifier* is intentionally **not** verified when a token is used, so tokens will remain valid if it is modified. This doesn't impact the security of our access tokens. For example, the following tokens have the same *token value* so are equivalent:

* `sgp_foobar_abcdef0123456789`
* `sgp_bazbar_abcdef0123456789`
